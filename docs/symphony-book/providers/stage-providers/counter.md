# Counter stage provider

Counter is a stateful stage provider. It can maintain states of multiple variables. You can add a new variable to the counter provider by declaring a new input field. The value of the input field decides how the variable is updated. For example, an input field `foo=1` defines a new variable `foo` with an initial value of `1`. And the variable is incremented by `1` every time the stage is executed. 

You can use the counter provider to count stage execution events. For example, if you have an HTTP stage provider named `http-stage` that sends a request to a given URL, you can use a counter provider to count number of successful requests and failed requests (see samples below).

## Inputs

| Field | Value |
|-------|-------|
| `<field>` | field increment |
| `<field>.init` | (optional) initial value of `<field>`. Default initial value is `0`. |

## Outputs

| Field | Value |
|-------|-------|
| `__state` | Variable states. You don't need to do anything with this value. Symphony automatically posts it back to the stage provider during next exeuction. This is how state is kept between execution. |

## Samples

Set up a counter named `foo` with initial value `1` and increment `1`. This configuration counts how many times the stage has been executed. To get the counter value, use the `$output()` function: `$output(foo-counter, foo)`.

```yaml
foo-counter:
  name: "foo-counter"
  provider: "providers.stage.counter"
  inputs:
    foo: 1      
```

Set up a counter `foo` with initial value `5` and increment `1`. After the first execution, `foo` becomes `6`.

```yaml
foo-counter:
  name: "foo-counter"
  provider: "providers.stage.counter"
  inputs:
    foo: 1
    foo.init: 5  
```

Count HTTP request results.

```yaml
http-stage:
name: "http-stage"
  provider: "providers.stage.http"
  inputs:
    url: "http://some/url"
    method: "GET"
  stageSelector: "counter"    
counter:
  name: "counter"
  provider: "providers.stage.counter"
  inputs:
    successes: ${{$if($equal($output(test,status),200),1,0)}}
    internal-errors: ${{$if($equal($output(test,status),500),1,0)}}
```

