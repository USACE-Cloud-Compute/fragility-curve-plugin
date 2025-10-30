# Single Sample Action Plugin

## Description
The single-sample action plugin computes a single fragility curve sample using specified seeds for natural variability and knowledge uncertainty. It takes a fragility curve model definition and event configuration with associated seeds, then generates a failure elevation sample for each probability point in the fragility curve.

## Implementation Details
The action implements the `cc.ActionRunner` interface and registers itself under the name "single-sample". It reads a fragility curve model from the "fragilitycurve" input data source and event configuration with seeds from the "seeds" input data source. The plugin uses the provided block seed and realization seed to compute a single sample of the fragility curve, generating failure elevations for each probability point.

## Process Flow
2. Read fragility curve model from "fragilitycurve" input data source
3. Read event-configuration with seeds from "seeds" input data source
4. Extract seed set named "fragilitycurveplugin" from event configuration
5. Compute single fragility curve sample using the extracted seeds
6. Marshal result to JSON format
7. Write computed result to output data source

## Configuration

### Environment
- Requires access to input data sources "fragilitycurve" and "seeds"
- Must have proper permissions to read from input data sources and write to output data source

### Attributes

### Action
No action attribute configuration required

### Global
No global attribute configuration required

### Input Configuration
 - **fragilitycurve** data source
 - **seeds** data source

### Input Data Sources
- **fragilitycurve**: JSON formatted fragility curve model containing probability-stage relationships with distribution parameters.  The data source requires a **Path** with a key of **default**
- **seeds**: JSON formatted event configuration containing seed sets for natural variability and knowledge uncertainty. The data source requires a **Path** with a key of **default**

### Output Configuration
 - output should be configured as a single data source with a **Path** key of **default**

### STORES
- No explicit store configuration required

## Configuration Examples

```json
{
  "name": "single-sample",
  "type": "single-sample",
  "description": "create failure_elevations",
  "inputs": [
    {
      "name": "fragilitycurve",
      "paths": {
        "default": "{ATTR::scenario}/system-response/conformance_system_response_curves.json"
      },
      "store_name": "FFRD"
    },
    {
      "name": "seeds",
      "paths": {
        "default": "{ATTR::scenario}/{ATTR::outputroot}/seeds.json"
      },
      "store_name": "FFRD"
    }
  ],
  "outputs": [
    {
      "name": "failure_elevations",
      "paths": {
        "default": "{ATTR::scenario}/{ATTR::outputroot}/system-response/failure_elevation.csv",
      },
      "store_name": "FFRD"
    }
  ]
}

```

## Outputs

### Format
JSON formatted output containing computed fragility curve location result (i.e. failure elevation)

### Data structures
```json
{
  "results": [
    {
      "location": "levee1",
      "failure_elevation": 1.00
    }
  ]
}
```

### Fields
- **location**: Name of the infrastructure location
- **failure_elevation**: the value of the failure elevation

## Error Handling
Errors are logged to the compute environment and processing will stop on error
