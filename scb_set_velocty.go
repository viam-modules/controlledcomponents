package controlledcomponents

import (
	"context"

	"github.com/golang/geo/r3"
)

// SetVelocity commands a base to move at the requested linear and angular velocites.
// When controls are enabled, SetVelocity polls the provided velocity movement sensor and corrects
// any error between the desired velocity and the actual velocity using a PID control loop.
func (sb *sensorBase) SetVelocity(
	ctx context.Context, linear, angular r3.Vector, extra map[string]interface{},
) error {
	sb.opMgr.CancelRunning(ctx)
	ctx, done := sb.opMgr.New(ctx)
	defer done()

	if sb.controlLoopConfig == nil {
		sb.logger.CWarnf(ctx, "control parameters not configured, using %v's SetVelocity method", sb.controlledBase.Name().ShortName())
		return sb.controlledBase.SetVelocity(ctx, linear, angular, extra)
	}

	// check tuning status
	if err := sb.checkTuningStatus(); err != nil {
		return err
	}

	// make sure the control loop is enabled
	if sb.loop == nil {
		if err := sb.startControlLoop(); err != nil {
			return err
		}
	}

	// convert linear.Y mmPerSec to mPerSec, angular.Z is degPerSec
	if err := sb.updateControlConfig(ctx, linear.Y/1000.0, angular.Z); err != nil {
		return err
	}
	sb.loop.Resume()

	return nil
}
