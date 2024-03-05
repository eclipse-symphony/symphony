# Defining a configuration object

In Symphony, a configuration object is described by a `catalog` object, which can hold a collection of key-value pairs, as shown in the following example:

```yaml
apiVersion: federation.symphony/v1
kind: Catalog
metadata:
  name: robot-config
spec: 
  type: config
  properties:
    name: my-robot
    os_version: "1.34"
    threshold: 30
    complex_object:
      foo1: bar1
      foo2: bar2
```

You can use [Symphony expressions](../concepts/unified-object-model/property-expressions.md) in configuration properties, such as referring to secrets.

Because Symphony allows arbitrary types in configuration values, you can embed your existing configuration objects without schema changes as embedded YAML or stringified JSON. Note that when you use complex object types, you can still use Symphony expressions at any levels. Symphony will resolve those expressions before it serves the configuration to the application.

> **NOTE:** Some applications use parametrized configurations that are managed by a separate system (not managed by Symphony). Symphony allows you to continue to use those parameters. This is why Symphony uses a specific syntax (`${{}})`) so that Symphony expressions can be distinguished from other parameterization systems. However, if such conflicts occur, you'll need to encode your parameter syntax so that it's not confused with Symphony expressions. 