package controlledcomponents

import (
	"context"
	"errors"

	"github.com/golang/geo/r3"
	"go.viam.com/rdk/components/base"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/spatialmath"
	"go.viam.com/utils/rpc"
)

var (
	SensorControlled = resource.NewModel("viam", "controlled-components", "sensor-controlled")
	errUnimplemented = errors.New("unimplemented")
)

func init() {
	resource.RegisterComponent(base.API, SensorControlled,
		resource.Registration[base.Base, *Config]{
			Constructor: newControlledComponentsSensorControlled,
		},
	)
}

type Config struct {
	/*
		Put config attributes here. There should be public/exported fields
		with a `json` parameter at the end of each attribute.

		Example config struct:
			type Config struct {
				Pin   string `json:"pin"`
				Board string `json:"board"`
				MinDeg *float64 `json:"min_angle_deg,omitempty"`
			}

		If your model does not need a config, replace *Config in the init
		function with resource.NoNativeConfig
	*/
}

// Validate ensures all parts of the config are valid and important fields exist.
// Returns implicit dependencies based on the config.
// The path is the JSON path in your robot's config (not the `Config` struct) to the
// resource being validated; e.g. "components.0".
func (cfg *Config) Validate(path string) ([]string, error) {
	// Add config validation code here
	return nil, nil
}

type controlledComponentsSensorControlled struct {
	resource.AlwaysRebuild

	name resource.Name

	logger logging.Logger
	cfg    *Config

	cancelCtx  context.Context
	cancelFunc func()
}

func newControlledComponentsSensorControlled(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (base.Base, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewSensorControlled(ctx, deps, rawConf.ResourceName(), conf, logger)

}

func NewSensorControlled(ctx context.Context, deps resource.Dependencies, name resource.Name, conf *Config, logger logging.Logger) (base.Base, error) {

	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	s := &controlledComponentsSensorControlled{
		name:       name,
		logger:     logger,
		cfg:        conf,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
	}
	return s, nil
}

func (s *controlledComponentsSensorControlled) Name() resource.Name {
	return s.name
}

func (s *controlledComponentsSensorControlled) NewClientFromConn(ctx context.Context, conn rpc.ClientConn, remoteName string, name resource.Name, logger logging.Logger) (base.Base, error) {
	panic("not implemented")
}

func (s *controlledComponentsSensorControlled) MoveStraight(ctx context.Context, distanceMm int, mmPerSec float64, extra map[string]interface{}) error {
	panic("not implemented")
}

func (s *controlledComponentsSensorControlled) Spin(ctx context.Context, angleDeg float64, degsPerSec float64, extra map[string]interface{}) error {
	panic("not implemented")
}

func (s *controlledComponentsSensorControlled) SetPower(ctx context.Context, linear r3.Vector, angular r3.Vector, extra map[string]interface{}) error {
	panic("not implemented")
}

func (s *controlledComponentsSensorControlled) SetVelocity(ctx context.Context, linear r3.Vector, angular r3.Vector, extra map[string]interface{}) error {
	panic("not implemented")
}

func (s *controlledComponentsSensorControlled) Stop(ctx context.Context, extra map[string]interface{}) error {
	panic("not implemented")
}

func (s *controlledComponentsSensorControlled) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	panic("not implemented")
}

func (s *controlledComponentsSensorControlled) IsMoving(ctx context.Context) (bool, error) {
	panic("not implemented")
}

func (s *controlledComponentsSensorControlled) Properties(ctx context.Context, extra map[string]interface{}) (base.Properties, error) {
	panic("not implemented")
}

func (s *controlledComponentsSensorControlled) Geometries(ctx context.Context, extra map[string]interface{}) ([]spatialmath.Geometry, error) {
	panic("not implemented")
}

func (s *controlledComponentsSensorControlled) Close(context.Context) error {
	// Put close code here
	s.cancelFunc()
	return nil
}
