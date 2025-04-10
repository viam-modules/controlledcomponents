# Module controlled-components 

This module houses components that utilize sensor feedback to transform sensor readings and/or control actuators.

## Model viam:controlled-components:sensor-controlled

The `sensor-controlled` model is a base that combines movement sensors with PID controls to actuate a `base` component.

### Configuration
The following attribute template can be used to configure this model:

```json
{
"base": <string>,
"movement_sensor": [<string>, <string>],
"control_frequency_hz": <float>,
"control_parameters": [
    {
        "type": "linear_velocity",
        "p": <float>,
        "i": <float>,
        "d": <float>,
    },
    {
        "type": "angular_velocity",
        "p": <float>,
        "i": <float>,
        "d": <float>
    },
  ]
}
```

The combination of movement sensors **must** provide the `AngularVelocity` and `LinearVelocity` endpoints. Providing the `Position`, `Orientation`, and `CompassHeading` endpoints will also improve the behavior of the base, but are not required.

#### Attributes

The following attributes are available for this model:

| Name          | Type   | Inclusion | Description                |
|---------------|--------|-----------|----------------------------|
| `base` | string | Required  | The name of the base that we want to apply PID controls to |
| `movement_sensor` | []string | Required  | the movement sensors that will be used for controls. The combination of movement sensors **must** provide the `AngularVelocity` and `LinearVelocity` endpoints. Providing the `Position`, `Orientation`, and `CompassHeading` endpoints will also improve the behavior of the base, but are not required. |
| `control_frequency_hz` | float64 | Optional  | the frequency that the PID controller will run at. Ensure this frequency is less than or equal to the movement sensor's supported frequency. **Default** is 10 Hz |
| `control_parameters` | []object  | Required  | an array of objects that provide the gains of the PID controller. Two must be configured. |

The control parameter object has the following parameters. Setting the PID gains to all be 0 will put the base in PID tuning mode.

| Name          | Type   | Inclusion | Description                |
|---------------|--------|-----------|----------------------------|
| `type` | string  | Required  | specifys what the PID values are controlling. Must be `linear_velocity` or `angular_velocity` |
| `p` | float  | Required  | the proportional gain for PID controls |
| `i` | float  | Required  | the proportional gain for PID controls |
| `d` | float  | Required  | the proportional gain for PID controls |

#### Example Configuration - Automatically tune the base

To configure your base to automatically tune, use the following configuration:

```json
{
"base": "my-wheeled-base",
"movement_sensor": ["my-wheeled-odometry"],
"control_parameters": [
    {
        "type": "linear_velocity",
        "p": 0,
        "i": 0,
        "d": 0,
    },
    {
        "type": "angular_velocity",
        "p": 0,
        "i": 0,
        "d": 0
    },
  ]
}
```

**WARNING**: Please have your base in a safe location, as it will begin moving once the machine finishes configuring.

After tuning is completed, update the PID values in your config. The PID values can be found in the machine's logs or via the DoCommand.

#### Example Configuration - Using tuned parameters

If you already have tuned PID values, the configuration should look like:

```json
{
"base": "my-wheeled-base",
"movement_sensor": ["my-wheeled-odometry"],
"control_parameters": [
    {
      "type": "linear_velocity",
      "p": 236.8709,
      "i": 708.7561,
      "d": 0
    },
    {
      "type": "angular_velocity",
      "p": 0.76573,
      "i": 2.30213,
      "d": 0
    }
  ]
}
```

### DoCommand

#### Get the Tuned PID gains of the base

This command will retrieve the tuned PID gains of the base when tuning has completed.

```json
{
  "get_tuned_pid": ""
}
```
