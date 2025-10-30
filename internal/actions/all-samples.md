# All Samples Action Plugin

## Description
The all-samples action plugin computes fragility curves using specified seeds for natural variability and knowledge uncertainty. It takes a fragility curve model definition and event configuration with associated seeds, then generates a failure elevation sample for each probability point in the fragility curve.

## Implementation Details
The action implements the `cc.ActionRunner` interface and registers itself under the name "all-samples". It reads a fragility curve model from the "fragilitycurve" input data source and event configuration with seeds from the "seeds" input data source. The plugin uses the provided block seed and realization seed to compute a set of samples of the fragility curve, generating failure elevations for each probability point.

## Process Flow
- Read fragility curve model from "fragilitycurve" input data source
- Read event-configuration with seeds from "seeds" input data source
- Extract seed set named "fragilitycurveplugin" from event configuration
- Compute fragility curves using the extracted seeds
- Marshal result to JSON format
- Write computed result to output data source

## Configuration

### Environment
- Requires access to input data sources "fragilitycurve" and "seeds"
- Must have proper permissions to read from input data sources and write to output data source

### Attributes

### Action
| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `seeds_format` | bool | No | When true seeds are read from an event store |
| `elevations_format` | bool | No | When true elevations are read from an event store |

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
 - at least one of two stores is required:
   - if an event store is being used, then a DataStore named 'store' and configured for tiledb, is required:
   ```json
    {
      "name": "store",
      "store_type": "TILEDB",
      "profile": "FFRD",
      "params": {
        "root": "model-library/ffrd-trinity/conformance/simulations"
      }
    }
    ```
   - if a file store is required, then a DataStore named 'seeds' is required:
   ```json
    {
      "name": "FFRD",
      "store_type": "S3",
      "profile": "FFRD",
      "params": {
        "root": "model-library/ffrd-trinity"
      }
    }
   ``` 

## Configuration Examples

```json
{
  "name": "all-samples",
  "type": "all-samples",
  "description": "create failure_elevations",
  "attributes": {
    "seeds_format": false,
    "elevations_format": false
  },
  "inputs": [
    {
      "name": "fragilitycurve",
      "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
      "paths": {
        "default": "{ATTR::scenario}/system-response/conformance_system_response_curves.json"
      },
      "store_name": "FFRD"
    },
    {
      "name": "seeds",
      "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
      "paths": {
        "default": "{ATTR::scenario}/{ATTR::outputroot}/seeds.json"
      },
      "store_name": "FFRD"
    }
  ],
  "outputs": [
    {
      "name": "failure_elevations",
      "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
      "paths": {
        "default": "{ATTR::scenario}/{ATTR::outputroot}/system-response/failure_elevations.csv",
        "event": "{ATTR::scenario}/{ATTR::outputroot}/event-data/${VAR::eventnumber}/system-response/failure_elevations.json"
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