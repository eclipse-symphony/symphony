# Delay stage provider

Delay stage provider sleeps for the given duration. It runs its `stageSelector` only after delay has expired.

## Inputs

| Field | Value |
|-------|-------|
| `delay` | an integer (in seconds) or a duration expression, such as `100` or `"4m20s"`

## Outputs

| Field | Value |
|-------|-------|
| `__status` | OK (200) |

## Sample

Sleep for 3 minutes before activating the `next-stage`:

```yaml
delay-stage:
  name: "delay-stage"
  provider: "providers.stage.delay"
  inputs:
    delay: "180s"
  stageSelector: "next-stage"
```

