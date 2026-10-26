# fragility-curve-plugin

**Repo class:** Plugin

A Cloud Compute plugin that samples system-response (fragility) curves to produce a failure elevation for each modeled location, for each event in a simulation.

> This documentation assumes familiarity with the FFRD SOPs and is intended for Federal employees and contractors. For Cloud Compute concepts (plugin, payload, data source, data store, action), see [cc-home](https://github.com/usace-cloud-compute/cc-home/tree/main/docs).

## What it does

A fragility curve relates the water surface elevation (stage) at a piece of infrastructure, such as a levee, to the probability that it fails. Each event in a simulation needs one failure elevation per location. The plugin gets it in two steps:

1. **Knowledge uncertainty.** If the curve's probabilities are uncertain (stored as distributions), the plugin draws one concrete curve from them using the uncertainty seed.
2. **Natural variability.** It draws a uniform random probability using the variability seed and reads the matching stage off the sampled curve. That stage is the location's failure elevation for the event.

Both draws come from seeds that an upstream seed generator (e.g. [seed-generator-plugin](https://github.com/usace-cloud-compute/seed-generator-plugin)) writes under the seed set name `fragilitycurveplugin`. Because results depend only on the seeds, a rerun with the same seeds gives the same elevations.

| Seed field | Used for |
|---|---|
| `block_seed` | Natural variability (which point on the curve) |
| `realization_seed` | Knowledge uncertainty (which curve) |
| `event_seed` | Not used |

## Actions

| Action | Use when | Details |
|---|---|---|
| `single-sample` | Running one event per plugin invocation | [single-sample.md](internal/actions/single-sample.md) |
| `all-samples` | Computing every event in one invocation, optionally to/from a TileDB event store | [all-samples.md](internal/actions/all-samples.md) |

Actions are independent; a payload normally contains one of them.

## Input: the fragility curve model

Both actions read a JSON model from the input data source named `fragilitycurve` (path key `default`).

```json
{
  "name": "example",
  "locations": [
    {
      "location": "levee1",
      "probability-stage": {
        "xvalues": [100, 102, 104, 106, 108],
        "ydistributions": [
          { "type": "DeterministicDistribution", "parameters": { "value": 0.0 } },
          { "type": "DeterministicDistribution", "parameters": { "value": 0.1 } },
          { "type": "DeterministicDistribution", "parameters": { "value": 0.5 } },
          { "type": "DeterministicDistribution", "parameters": { "value": 0.9 } },
          { "type": "DeterministicDistribution", "parameters": { "value": 1.0 } }
        ]
      }
    }
  ]
}
```

| Field | Meaning |
|---|---|
| `name` | Model name (informational) |
| `locations[].location` | Location name; it is copied to the output |
| `probability-stage.xvalues` | **Stages** (elevations), ascending |
| `probability-stage.ydistributions` | **Probability of failure** at each stage. Must increase with stage, ideally from 0 to 1. Use `DeterministicDistribution` for a fixed curve or another [go-statistics](https://github.com/HydrologicEngineeringCenter/go-statistics) distribution (`TriangularDistribution`, `NormalDistribution`, `UniformDistribution`, …) to represent knowledge uncertainty |

Optional descriptive fields per location, carried through but not used in the computation: `hydraulic_model_unit`, `breach_fid`, `nld_system_id`, `nld_system_name`, `nld_segment_id`, `nidid`, `nid_dam_name`, `failure_mode`, `source`, `top_elev`, `toe_elev`.

> **Axis order matters.** `xvalues` are stages and `ydistributions` are probabilities. Earlier versions of this README, and [configs/fragilitycurve.json](configs/fragilitycurve.json), have them the other way around (probability on x, elevation distributions on y). The plugin doesn't reject that file, but every failure elevation it produces from it is `0`.

If a sampled probability falls below the curve's lowest probability, the plugin returns a failure elevation of `0`, not the lowest stage. Start the curve at probability 0 to avoid that.

![Example fragility curve with uncertainty](example_fragility_curve.png)

*The figure illustrates the idea of an uncertain fragility curve. Its axes are drawn probability-vs-elevation; in the input JSON, the stage goes on x.*

## Outputs

| Action | Output |
|---|---|
| `single-sample` | One JSON file: `{"results":[{"location":"levee1","failure_elevation":104.3}]}` |
| `all-samples` (file mode) | A CSV of every event (`event_number,<location>,…`) plus one JSON file per event |
| `all-samples` (TileDB mode) | A dense TileDB array `failure_elevations` (events × locations) in the store named `store` |

See each action's doc for path keys and layout.

## Payload attributes

| Attribute | Action | Type | Default | Description |
|---|---|---|---|---|
| `seeds_format` | `all-samples` | bool | `false` | `true`: read seeds from the TileDB store named `store` |
| `elevations_format` | `all-samples` | bool | `false` | `true`: write failure elevations to the TileDB store named `store` |

`single-sample` takes no attributes.

## Environment variables

The plugin defines no environment variables of its own. It uses the standard Cloud Compute plugin environment that [cc-go-sdk](https://github.com/usace-cloud-compute/cc-go-sdk) reads (payload location, store credentials and profiles); see the cc-home docs.

## Container image

```
ghcr.io/usace-cloud-compute/fragility-curve-plugin
```

The image is built and pushed by [.github/workflows/build-push-ghcr.yml](.github/workflows/build-push-ghcr.yml). The binary is installed at `/app/fragility-curves`. The repo doesn't contain a plugin manifest yet; when you register the plugin, set the command to `/app/fragility-curves` explicitly.

**Resources:** not yet measured. TODO: record vCPU and memory from a representative run (STD-PLG-030). Memory grows with (number of events × number of locations) in `all-samples`, which holds all results in memory before writing. Each invocation runs on one thread, and events are embarrassingly parallel across `single-sample` invocations.

**Network:** the plugin only connects to the data stores declared in the payload.

## Build and test

Requires Go 1.23+ and the TileDB C library (CGO). The easiest route is the dev container ([.devcontainer/devcontainer.json](.devcontainer/devcontainer.json)), which uses the `dev` stage of the [Dockerfile](Dockerfile).

```sh
make build    # builds ./fragility-curves, stamping commit and build date
make test     # go test ./...
docker build -t fragility-curve-plugin .
```

Note: the tests in [internal/compute/model_test.go](internal/compute/model_test.go) read and write data under `/workspaces/fragilitycurveplugin/configs/`, and most of those files aren't in the repo. Outside the original dev environment, most tests fail.

[docker-compose.yml](docker-compose.yml) starts local MinIO (S3), Redis and an SQS emulator for development. The credentials in it are local placeholders.

## Repository layout

| Path | Contents |
|---|---|
| `main.go` | Entry point: registers the TileDB store type and runs the payload's actions |
| `internal/actions/` | `single-sample` and `all-samples` action implementations and their docs |
| `internal/compute/` | Fragility curve model, sampling, seed reading and TileDB output |
| `configs/` | Sample input files (see the axis-order note above) |

`configs/levee1_fragilitycurve.csv` is not read by the plugin. It shows the per-location CSV format that test helpers convert into the JSON model.

## License

MIT. See [LICENSE](LICENSE).
