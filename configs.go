package controlledcomponents

import (
	"github.com/pkg/errors"

	"go.viam.com/rdk/control"
	"go.viam.com/rdk/resource"
)

var (
	family = resource.NewModelFamily("viam", "controlled-components")
	// SensorControlledModel is the name of the sensor_controlled model of a base component.
	SensorControlledModel = family.WithModel("sensor-controlled")
)

// SCBConfig configures a sensor controlled base.
type SCBConfig struct {
	MovementSensor    []string            `json:"movement_sensor"`
	Base              string              `json:"base"`
	ControlParameters []control.PIDConfig `json:"control_parameters,omitempty"`
	ControlFreq       float64             `json:"control_frequency_hz,omitempty"`
}

// Validate validates all parts of the sensor controlled base config.
func (cfg *SCBConfig) Validate(path string) ([]string, error) {
	deps := []string{}
	if len(cfg.MovementSensor) == 0 {
		return nil, resource.NewConfigValidationError(path, errors.New("need at least one movement sensor for base"))
	}
	deps = append(deps, cfg.MovementSensor...)

	if cfg.Base == "" {
		return nil, resource.NewConfigValidationFieldRequiredError(path, "base")
	}
	deps = append(deps, cfg.Base)

	for _, pidConf := range cfg.ControlParameters {
		if pidConf.Type != typeLinVel && pidConf.Type != typeAngVel {
			return nil, resource.NewConfigValidationError(path,
				errors.New("control_parameters type must be 'linear_velocity' or 'angular_velocity'"))
		}
	}

	return deps, nil
}
