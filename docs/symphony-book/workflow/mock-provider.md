# Mock Stage Provider

Mock stage provider is a simple provider for testing. It replays all `inputs` in its `outputs` with one exception: If your `inputs` contains a `foo` field with an integer value, it will increment the field by `1`. For example, if you have input `foo=100`, you'll get output `foo=101`.

A sample state with a mock provider:

```yaml
stages:
mock:
    name: "mock"
    provider: "providers.stage.mock"   
```

## Use mock stage provider to construct a loop

In the following stage definition, the ouput field `foo` is fed back to the stage input. And becaue the mock stage provider increments the `foo` value, the output `foo` will be incremented by `1` at each iteration. The stage selector checks the output value and select the `mock` stage again if the value of `foo` is less than `5`. Otherwise, the selector selects an empty stage, which stops the workflow execution. 

```yaml
mock:
    name: "mock"
    provider: "providers.stage.mock"
    inputs:
      foo: "${{$output(mock,foo)}}"
    stageSelector: "${{$if($lt($output(mock,foo), 5), mock, '')}}"
```