// Package controlledcomponents implements a components that will be controlled using components
// this code implements a sensor-controlled base with feedback control from a movement sensor
package controlledcomponents

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang/geo/r3"
	"github.com/pkg/errors"
	"go.viam.com/rdk/components/base"
	"go.viam.com/rdk/components/movementsensor"
	"go.viam.com/rdk/control"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/operation"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/spatialmath"
)

const (
	yawPollTime        = 5 * time.Millisecond
	velocitiesPollTime = 5 * time.Millisecond
	sensorDebug        = false
	typeLinVel         = "linear_velocity"
	typeAngVel         = "angular_velocity"
	defaultControlFreq = 10 // Hz
	getPID             = "get_tuned_pid"
)

var errNoGoodSensor = errors.New("no appropriate sensor for orientation or velocity feedback")

func init() {
	resource.RegisterComponent(
		base.API,
		SensorControlledModel,
		resource.Registration[base.Base, *SCBConfig]{Constructor: newSCB})
}

type sensorBase struct {
	name   resource.Name
	conf   *SCBConfig
	logger logging.Logger
	mu     sync.Mutex

	activeBackgroundWorkers sync.WaitGroup
	controlledBase          base.Base // the inherited wheeled base

	opMgr *operation.SingleOperationManager

	allSensors []movementsensor.MovementSensor
	velocities movementsensor.MovementSensor
	position   movementsensor.MovementSensor
	// headingFunc returns the current angle between (-180,180) and whether Spin is supported
	headingFunc func(ctx context.Context) (float64, bool, error)

	controlLoopConfig *control.Config
	blockNames        map[string][]string
	loop              *control.Loop
	configPIDVals     []control.PIDConfig
	tunedVals         *[]control.PIDConfig
	controlFreq       float64
}

func newSCB(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (base.Base, error) {
	conf, err := resource.NativeConfig[*SCBConfig](rawConf)
	if err != nil {
		return nil, err
	}

	return NewSensorControlled(ctx, deps, rawConf.ResourceName(), conf, logger)
}

// NewSensorControlled creates a new sensorcontrolled base using the base API.
func NewSensorControlled(ctx context.Context, deps resource.Dependencies,
	name resource.Name, conf *SCBConfig, logger logging.Logger,
) (base.Base, error) {
	sb := &sensorBase{
		logger:        logger,
		tunedVals:     &[]control.PIDConfig{{}, {}},
		configPIDVals: []control.PIDConfig{{}, {}},
		name:          name,
		opMgr:         operation.NewSingleOperationManager(),
	}

	if err := sb.reconfigureWithConfig(ctx, deps, conf); err != nil {
		return nil, err
	}

	return sb, nil
}

func (sb *sensorBase) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	newConf, err := resource.NativeConfig[*SCBConfig](conf)
	if err != nil {
		return err
	}

	return sb.reconfigureWithConfig(ctx, deps, newConf)
}

func (sb *sensorBase) reconfigureWithConfig(ctx context.Context, deps resource.Dependencies, newConf *SCBConfig) error {
	var err error
	if sb.loop != nil {
		sb.loop.Stop()
		sb.loop = nil
	}

	sb.mu.Lock()
	defer sb.mu.Unlock()

	sb.controlFreq = defaultControlFreq
	if newConf.ControlFreq != 0 {
		sb.controlFreq = newConf.ControlFreq
	}

	// reset all sensors
	sb.allSensors = nil
	sb.velocities = nil
	var orientation movementsensor.MovementSensor
	var compassHeading movementsensor.MovementSensor
	sb.position = nil
	sb.controlledBase = nil

	for _, name := range newConf.MovementSensor {
		ms, err := movementsensor.FromDependencies(deps, name)
		if err != nil {
			return errors.Wrapf(err, "no movement sensor named (%s)", name)
		}
		sb.allSensors = append(sb.allSensors, ms)
	}

	for _, ms := range sb.allSensors {
		props, err := ms.Properties(context.Background(), nil)
		if err == nil && props.OrientationSupported {
			// return first sensor that does not error that satisfies the properties wanted
			orientation = ms
			sb.logger.CInfof(ctx, "using sensor %s as orientation sensor for base", orientation.Name().ShortName())
			break
		}
	}

	for _, ms := range sb.allSensors {
		props, err := ms.Properties(context.Background(), nil)
		if err == nil && props.AngularVelocitySupported && props.LinearVelocitySupported {
			// return first sensor that does not error that satisfies the properties wanted
			sb.velocities = ms
			sb.logger.CInfof(ctx, "using sensor %s as velocity sensor for base", sb.velocities.Name().ShortName())
			break
		}
	}

	for _, ms := range sb.allSensors {
		props, err := ms.Properties(context.Background(), nil)
		if err == nil && props.PositionSupported {
			// return first sensor that does not error that satisfies the properties wanted
			sb.position = ms
			sb.logger.CInfof(ctx, "using sensor %s as position sensor for base", sb.position.Name().ShortName())
			break
		}
	}

	for _, ms := range sb.allSensors {
		props, err := ms.Properties(context.Background(), nil)
		if err == nil && props.CompassHeadingSupported {
			// return first sensor that does not error that satisfies the properties wanted
			compassHeading = ms
			sb.logger.CInfof(ctx, "using sensor %s as compassHeading sensor for base", compassHeading.Name().ShortName())
			break
		}
	}
	sb.determineHeadingFunc(ctx, orientation, compassHeading)

	if orientation == nil && sb.velocities == nil {
		return errNoGoodSensor
	}

	sb.controlledBase, err = base.FromDependencies(deps, newConf.Base)
	if err != nil {
		return errors.Wrapf(err, "no base named (%s)", newConf.Base)
	}

	if sb.velocities != nil && len(newConf.ControlParameters) != 0 {
		// assign linear and angular PID correctly based on the given type
		for _, pidConf := range newConf.ControlParameters {
			switch pidConf.Type {
			case typeLinVel:
				// configPIDVals at index 0 is linear
				sb.configPIDVals[0] = pidConf
			case typeAngVel:
				// configPIDVals at index 1 is angular
				sb.configPIDVals[1] = pidConf
			default:
				return fmt.Errorf("control_parameters type '%v' not accepted, type must be 'linear_velocity' or 'angular_velocity'",
					pidConf.Type)
			}
		}

		// unlock the mutex before setting up the control loop so that the motors
		// are not locked, and can run if any auto-tuning is necessary
		sb.mu.Unlock()
		if err := sb.setupControlLoop(sb.configPIDVals[0], sb.configPIDVals[1]); err != nil {
			sb.mu.Lock()
			return err
		}
		// relock the mutex after setting up the control loop since there is still a defer unlock
		sb.mu.Lock()
	}
	sb.conf = newConf

	return nil
}

func (sb *sensorBase) Name() resource.Name {
	return sb.name
}

func (sb *sensorBase) SetPower(
	ctx context.Context, linear, angular r3.Vector, extra map[string]interface{},
) error {
	sb.opMgr.CancelRunning(ctx)
	if sb.loop != nil {
		sb.loop.Pause()
	}
	return sb.controlledBase.SetPower(ctx, linear, angular, extra)
}

func (sb *sensorBase) Stop(ctx context.Context, extra map[string]interface{}) error {
	sb.opMgr.CancelRunning(ctx)
	if sb.loop != nil {
		sb.loop.Pause()
		// update pid controllers to be an at rest state
		if err := sb.updateControlConfig(ctx, 0, 0); err != nil {
			return err
		}
	}
	return sb.controlledBase.Stop(ctx, extra)
}

func (sb *sensorBase) IsMoving(ctx context.Context) (bool, error) {
	return sb.controlledBase.IsMoving(ctx)
}

func (sb *sensorBase) Properties(ctx context.Context, extra map[string]interface{}) (base.Properties, error) {
	return sb.controlledBase.Properties(ctx, extra)
}

func (sb *sensorBase) Geometries(ctx context.Context, extra map[string]interface{}) ([]spatialmath.Geometry, error) {
	return sb.controlledBase.Geometries(ctx, extra)
}

func (sb *sensorBase) DoCommand(ctx context.Context, req map[string]interface{}) (map[string]interface{}, error) {
	resp := make(map[string]interface{})

	sb.mu.Lock()
	defer sb.mu.Unlock()

	if _, ok := req[getPID]; ok {
		controlParams := []control.PIDConfig{}
		for _, pidConf := range *sb.tunedVals {
			if !pidConf.NeedsAutoTuning() {
				controlParams = append(controlParams, pidConf)
			}
		}
		resp["control_parameters"] = controlParams
	}

	return resp, nil
}

func (sb *sensorBase) Close(ctx context.Context) error {
	if err := sb.Stop(ctx, nil); err != nil {
		return err
	}
	if sb.loop != nil {
		sb.loop.Stop()
		sb.loop = nil
	}

	sb.activeBackgroundWorkers.Wait()
	return nil
}
