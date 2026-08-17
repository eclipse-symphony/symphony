# Mock stage provider

Mock stage provider echoes its inputs back as outputs. It also prints the inputs and outputs to the log, which makes it useful for testing campaign flows and for building sample campaigns. As a special case, an input named `foo` is treated as a counter: its integer value is incremented by `1` on every execution (starting from `1` when not set), so a mock stage can loop a configurable number of times when combined with a `stageSelector` expression.

## Inputs

| Field | Value |
|-------|-------|
| `<any field>` | Any input field is copied to the outputs as is. |
| `foo` | (optional) An integer counter that is incremented by `1` on each execution. |

## Outputs

| Field | Value |
|-------|-------|
| `<any field>` | A copy of each input field. |
| `foo` | The incremented counter value, if `foo` was present in the inputs. |

## Sample

Loop the `mock` stage until the `foo` counter reaches `5`:

```yaml
mock:
  name: "mock"
  provider: "providers.stage.mock"
  inputs:
    foo: "${{$output(mock,foo)}}"
  stageSelector: "${{$if($lt($output(mock,foo), 5), mock, '')}}"
```
