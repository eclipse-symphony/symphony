# AI skill

`skill.ai.symphony` defines a processing graph that is made up of `edges` and `nodes`.

## Schema

| Field | Type | Description |
|--------|--------|--------|
| `Bindings`| `[]BindingSpec` | A list of binding.md that represent actions allowed on the AI skill |
| `DisplayName` | `string` | A user friendly name |
| `Edges` | `[]EdgeSpec` | graph edges |
| `Nodes` | `[]NodeSpec` | graph nodes |
| `Parameters` | `map[string]string` | Parameters. A parameter can be used anywhere in the skill definition. See the [parameters](#parameters) sections below |
| `Properties` | `map[string]string` | A property bag |

## Parameters

You can define any number of parameters in the `parameters` section. All parameters are `string` typed. When you declare a parameter, you specify a default value.

```yaml
parameters:   
  delay_buffer: "0.1"
  model_platform: "invalid"
  model_flavor: "edge"
```

Then, you can refer to the parameter in your AI skill graph definition, for example:

```yaml
nodes:
- id: "4"
  type: export
  name: video_snippet_export
  configurations:
    filename_prefix: test
    recording_duration: "$param(model_flavor)"
    insights_overlay: "$param(model_platform)"
    delay_buffer: "$param(delay_buffer)"  
```

## Overwrite parameters

An `instance.solution.symphony` object can overwrite AI skill parameter values in its own `parameters` section.

```yaml
parameters:
  cv-skill.delay_buffer: "0.2"
  cv-skill.model_platform: "TensorFlow"
  cv-skill.model_flavor: "TensorFlowMobile"
```

When a solution instance uses multiple references to a same AI skill, it can use aliases to distinguish between different AI skill instances.
