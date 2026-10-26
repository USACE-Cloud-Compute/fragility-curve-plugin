# Single Sample Action

## Description
The `single-sample` action computes the failure elevation at every location in a fragility curve model for **one event**. It reads the model and that event's seeds, samples each curve once, and writes one JSON result.

Use it when each event runs as its own plugin invocation. To compute many events in one invocation, use [all-samples](all-samples.md).

## Process Flow
1. Read the fragility curve model from the `fragilitycurve` input data source.
2. Read the event configuration from the `seeds` input data source.
3. Take the seed set named `fragilitycurveplugin`. If it is missing, the action fails.
4. For each location, sample the curve using `realization_seed` (knowledge uncertainty), then sample a failure elevation using `block_seed` (natural variability).
5. Write the result as JSON to the single output data source's `default` path.

## Configuration

### Attributes
None.

### Inputs
| Data source name | Path key | Format |
|---|---|---|
| `fragilitycurve` | `default` | Fragility curve model JSON (see the [README](../../README.md#input-the-fragility-curve-model)) |
| `seeds` | `default` | A single event configuration object (below) |

Seeds file:
```json
{
  "seeds": {
    "fragilitycurveplugin": {
      "event_seed": 1234,
      "block_seed": 5678,
      "realization_seed": 9012
    }
  }
}
```
Other seed sets in the file are ignored.

### Outputs
Exactly one output data source is required, with a `default` path key. Its name is up to you.

### Stores
Any store type supported by cc-go-sdk that the input and output data sources reference.

## Configuration Example

```json
{
  "name": "single-sample",
  "type": "single-sample",
  "description": "create failure elevations",
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
        "default": "{ATTR::scenario}/{ATTR::outputroot}/system-response/failure_elevations.json"
      },
      "store_name": "FFRD"
    }
  ]
}
```

## Output

```json
{
  "results": [
    { "location": "levee1", "failure_elevation": 104.3 }
  ]
}
```

- **location**: the location name from the model
- **failure_elevation**: the sampled stage at which the location fails for this event, in the model's elevation units

Results are listed in the same order as the locations in the model.

## Error Handling
The action returns an error and stops if: the payload doesn't define exactly one output, an input can't be read or parsed, the `fragilitycurveplugin` seed set is missing, or the output can't be written. This action returns errors without logging them first. `main.go` logs the error and the process exits 0.
