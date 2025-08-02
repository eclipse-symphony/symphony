# Delay Stage Provider

The delay stage provider pause workflow execution for the given period of time, specified either as an integer in seconds or a [duration expression](https://pkg.go.dev/maze.io/x/duration#ParseDuration) such as `2h45m` (2 hours and 45 minutes).

The following sample delay stage delays for 30 seconds:

```yaml
delay:
  name: delay
  provider: providers.stage.delay
  inputs:
    delay: 30s
```
