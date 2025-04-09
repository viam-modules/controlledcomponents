package controlledcomponents

import (
	"context"
	"math"

	"go.viam.com/rdk/components/movementsensor"
	rdkutils "go.viam.com/rdk/utils"
)

func sign(x float64) float64 { // A quick helper function
	if math.Signbit(x) {
		return -1.0
	}
	return 1.0
}

// determineHeadingFunc determines which movement sensor endpoint should be used for control.
// The priority is Orientation -> Heading -> No heading control.
func (sb *sensorBase) determineHeadingFunc(ctx context.Context,
	orientation, compassHeading movementsensor.MovementSensor,
) {
	switch {
	case orientation != nil:

		sb.logger.CInfof(ctx, "using sensor %s as angular heading sensor for base %v", orientation.Name().ShortName(), sb.Name().ShortName())

		sb.headingFunc = func(ctx context.Context) (float64, bool, error) {
			orient, err := orientation.Orientation(ctx, nil)
			if err != nil {
				return 0, false, err
			}
			// this returns (-180-> 180)
			yaw := rdkutils.RadToDeg(orient.EulerAngles().Yaw)

			return yaw, true, nil
		}
	case compassHeading != nil:
		sb.logger.CInfof(ctx, "using sensor %s as angular heading sensor for base %v", compassHeading.Name().ShortName(), sb.Name().ShortName())

		sb.headingFunc = func(ctx context.Context) (float64, bool, error) {
			compass, err := compassHeading.CompassHeading(ctx, nil)
			if err != nil {
				return 0, false, err
			}
			// flip compass heading to be CCW/Z up
			compass = 360 - compass

			// make the compass heading (-180->180)
			if compass > 180 {
				compass -= 360
			}

			return compass, true, nil
		}
	default:
		sb.logger.CInfof(ctx, "base %v cannot control heading, no heading related sensor given",
			sb.Name().ShortName())
		sb.headingFunc = func(ctx context.Context) (float64, bool, error) {
			return 0, false, nil
		}
	}
}
