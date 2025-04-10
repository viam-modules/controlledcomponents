package controlledcomponents

import (
	"context"

	"github.com/golang/geo/r3"
	"go.viam.com/rdk/control"
)

// startControlLoop uses the control config to initialize a control loop and store it on the sensor controlled base struct.
// The sensor base is the controllable interface that implements State and GetState called from the endpoint block of the control loop.
func (sb *sensorBase) startControlLoop() error {
	loop, err := control.NewLoop(sb.logger, *sb.controlLoopConfig, sb)
	if err != nil {
		return err
	}
	if err := loop.Start(); err != nil {
		return err
	}
	sb.loop = loop

	return nil
}

func (sb *sensorBase) setupControlLoop(linear, angular control.PIDConfig) error {
	// set the necessary options for a sensorcontrolled base
	options := control.Options{
		SensorFeedback2DVelocityControl: true,
		LoopFrequency:                   sb.controlFreq,
		ControllableType:                "base_name",
	}

	// check if either linear or angular need to be tuned
	if linear.NeedsAutoTuning() || angular.NeedsAutoTuning() {
		options.NeedsAutoTuning = true
	}

	// combine linear and angular back into one control.PIDConfig, with linear first
	pidVals := []control.PIDConfig{linear, angular}

	// fully set up the control config based on the provided options
	pl, err := control.SetupPIDControlConfig(pidVals, sb.Name().ShortName(), options, sb, sb.logger)
	if err != nil {
		return err
	}

	sb.controlLoopConfig = pl.ControlConf
	sb.loop = pl.ControlLoop
	sb.blockNames = pl.BlockNames
	sb.tunedVals = pl.TunedVals

	return nil
}

func (sb *sensorBase) updateControlConfig(
	ctx context.Context, linearValue, angularValue float64,
) error {
	// set linear setpoint config
	if err := control.UpdateConstantBlock(ctx, sb.blockNames[control.BlockNameConstant][0], linearValue, sb.loop); err != nil {
		return err
	}

	// set angular setpoint config
	if err := control.UpdateConstantBlock(ctx, sb.blockNames[control.BlockNameConstant][1], angularValue, sb.loop); err != nil {
		return err
	}

	return nil
}

// SetState is called in endpoint.go of the controls package by the control loop
// instantiated in this file. It is a helper function to call the sensor-controlled base's
// SetVelocity from within that package.
func (sb *sensorBase) SetState(ctx context.Context, state []*control.Signal) error {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.loop != nil && !sb.loop.Running() {
		return nil
	}

	sb.logger.CDebug(ctx, "setting state")
	linvel := state[0].GetSignalValueAt(0)
	// multiply by the direction of the linear velocity so that angular direction
	// (cw/ccw) doesn't switch when the base is moving backwards
	angvel := (state[1].GetSignalValueAt(0) * sign(linvel))

	return sb.controlledBase.SetPower(ctx, r3.Vector{Y: linvel}, r3.Vector{Z: angvel}, nil)
}

// State is called in endpoint.go of the controls package by the control loop
// instantiated in this file. It is a helper function to call the sensor-controlled base's
// movementsensor and insert its LinearVelocity and AngularVelocity values
// in the signal in the control loop's thread in the endpoint code.
func (sb *sensorBase) State(ctx context.Context) ([]float64, error) {
	sb.logger.CDebug(ctx, "getting state")
	linvel, err := sb.velocities.LinearVelocity(ctx, nil)
	if err != nil {
		return []float64{}, err
	}

	angvel, err := sb.velocities.AngularVelocity(ctx, nil)
	if err != nil {
		return []float64{}, err
	}
	return []float64{linvel.Y, angvel.Z}, nil
}

// if loop is tuning, return an error
// if loop has been tuned but the values haven't been added to the config, error with tuned values.
func (sb *sensorBase) checkTuningStatus() error {
	done := true
	needsTuning := false

	for i := range sb.configPIDVals {
		// check if the current signal needed tuning
		if sb.configPIDVals[i].NeedsAutoTuning() {
			// return true if either signal needed tuning
			needsTuning = needsTuning || true
			// if the tunedVals have not been updated, then tuning is still in progress
			done = done && !(*sb.tunedVals)[i].NeedsAutoTuning()
		}
	}

	if needsTuning {
		if done {
			return control.TunedPIDErr(sb.Name().ShortName(), *sb.tunedVals)
		}
		return control.TuningInProgressErr(sb.Name().ShortName())
	}

	return nil
}
