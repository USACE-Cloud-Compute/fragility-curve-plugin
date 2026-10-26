# All Samples Action

## Description
The `all-samples` action computes the failure elevation at every location in a fragility curve model for **every event** in one invocation. It reads one seed set per event, samples each curve once per event, and writes the results either as files (CSV and per-event JSON) or as a TileDB array.

Use it to process a whole simulation at once. For one event per invocation, use [single-sample](single-sample.md).

## Process Flow
1. Read the fragility curve model from the `fragilitycurve` input data source.
2. Read seeds, one set per event, in order:
   - `seeds_format: false`: from the `seeds` input data source (JSON array)
   - `seeds_format: true`: from the TileDB array `seeds` in the store named `store`
3. For each event, sample each location's curve using `realization_seed` (knowledge uncertainty), then a failure elevation using `block_seed` (natural variability).
4. Write results:
   - `elevations_format: false`: a CSV of all events to the output's `default` path, and one JSON file per event to its `event` path
   - `elevations_format: true`: the TileDB array `failure_elevations` in the store named `store`

## Configuration

### Attributes
| Attribute | Type | Required | Default | Description |
|---|---|---|---|---|
| `seeds_format` | bool | No | `false` | `true`: read seeds from the TileDB event store instead of the `seeds` JSON input |
| `elevations_format` | bool | No | `false` | `true`: write failure elevations to the TileDB event store instead of CSV/JSON files |

### Inputs
| Data source name | Path key | Required when | Format |
|---|---|---|---|
| `fragilitycurve` | `default` | Always | Fragility curve model JSON (see the [README](../../README.md#input-the-fragility-curve-model)) |
| `seeds` | `default` | `seeds_format` is `false` | JSON **array** of event configurations, one per event, in event order |

Seeds file (JSON mode):
```json
[
  { "seeds": { "fragilitycurveplugin": { "event_seed": 11, "block_seed": 12, "realization_seed": 13 } } },
  { "seeds": { "fragilitycurveplugin": { "event_seed": 21, "block_seed": 22, "realization_seed": 23 } } }
]
```

### Outputs
Exactly one output data source must be defined, even in TileDB mode, where it isn't written to. In file mode it needs two path keys:

| Path key | Contents |
|---|---|
| `default` | CSV of all events |
| `event` | Per-event JSON. Must contain `${VAR::eventnumber}` so each event gets its own file; otherwise every event overwrites the same file |

### Stores
- **File mode:** any store type that the input and output data sources reference.
- **TileDB mode** (either attribute `true`): a store named exactly `store` with type `TILEDB`:
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
  - Reading seeds expects a dense array `seeds` (rows = events, columns = plugins) with int64 attributes `realization_seed`, `block_seed` and `event_seed`, and a `seed_columns` metadata entry listing the column names. One of those names must be `fragilitycurveplugin`.
  - Writing creates a dense array `failure_elevations` with dimensions `Events` (1…number of events) × `fragility_locations` (1…number of locations), a float64 attribute `failure_elevation`, and a `fragility_locations` metadata entry listing location names in column order.

  The names `store`, `seeds` and `failure_elevations` are hard-coded.

## Configuration Example (file mode)

```json
{
  "name": "all-samples",
  "type": "all-samples",
  "description": "create failure elevations",
  "attributes": {
    "seeds_format": false,
    "elevations_format": false
  },
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
        "default": "{ATTR::scenario}/{ATTR::outputroot}/system-response/failure_elevations.csv",
        "event": "{ATTR::scenario}/{ATTR::outputroot}/event-data/${VAR::eventnumber}/system-response/failure_elevations.json"
      },
      "store_name": "FFRD"
    }
  ]
}
```

## Outputs (file mode)

CSV (`default` path). There is one row per event and one column per location, in model order:
```
event_number,levee1,levee2
1,104.3,98.7
2,101.9,99.2
```
`event_number` is the 1-based position of the event in the seeds list, not the seed's `event_seed`.

Per-event JSON (`event` path, with `eventnumber` set to the same 1-based number):
```json
{
  "results": [
    { "location": "levee1", "failure_elevation": 104.3 },
    { "location": "levee2", "failure_elevation": 98.7 }
  ]
}
```

## Error Handling
The action logs the failure and stops on any error: the payload doesn't define exactly one output, an input or store can't be read, or an output can't be written. Progress is logged roughly every 10% of events.
