# Parameter management

Symphony allows `skill` objects to define parameters, which can be overwritten by an `instance` object during deployment. Symphony also allows an instance to create multiple sets of parameters when it uses multiple instances of the skill.

## Define and use parameters

You can define parameters on your AI skill objects. To declare a parameter, add `<parameter name>: <default value>` to the `parameters` collection of your skill object:

```yaml
apiVersion: ai.symphony/v1
kind: Skill
metadata:
  name: cv-skill
  labels: 
    foo: bar
spec:
  parameters:   
    delay_buffer: "0.1"
    insights_overlay: "yes"
    recording_duration: "3"
...
```

Then, in your skill definition, you can refer to a parameter using the `$param()` function, such as:

```yaml
configurations:
      filename_prefix: test
      recording_duration: "$param(recording_duration)"
      insights_overlay: "$param(insights_overlay)"
      delay_buffer: "$param(delay_buffer)" 
```

## Overwrite parameters

An `instance` can define multiple overrides of a parameter, in the format of `<AI skill name>.<parameter name>` or `<AI skill name>.<alias>.<parameter name>`.

For example, the following instance defines three sets of parameters of `cv-skill`. Two of them are aliased as `abc` and `def`, respectively.

```yaml
apiVersion: solution.symphony/v1
kind: Instance
metadata:
  name: dummy-instance
spec:
  parameters:
    cv-skill.delay_buffer: "0.2"
    cv-skill.insights_overlay: "yes"
    cv-skill.recording_duration: "3"
    cv-skill.abc.delay_buffer: "0.4"
    cv-skill.abc.insights_overlay: "false"
    cv-skill.abc.recording_duration: "6"
    cv-skill.def.delay_buffer: "0.5"
    cv-skill.def.insights_overlay: "yes"
    cv-skill.def.recording_duration: "34"
```

When an application queries AI skill through [Symphony agent](../agent/_overview.md), the following rules apply:

1. If the `instance` query parameter is missing, the default values from the skill definition are used.
2. If the `alias` query parameter is missing, the `<AI skill name>.<parameter name>` overrides from the instance object are used.
3. If the `alias` query parameter is present, the `<AI skill name>.<alias>.<parameter name>` overrides from the instance object are used.
4. If any overrides are missing, the default values from the skill definition are used.

## Query AI skills

See [Symphony agent](../agent/_overview.md) for details on how to query AI skills with parameter overwrites.
