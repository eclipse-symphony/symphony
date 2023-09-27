# Defining Configurations

In Symphony, a configuration object is described by a Catalog object, which can hold a collection of key-value pairs, as shown in the following example:

```yaml
apiVersion: federation.symphony/v1
kind: Catalog
metadata:
  name: robot-config
spec:  
  siteId: hq
  type: config
  name: robot-config
  properties:
    name: my-robot
    os_version: "1.34"
    threshold: 30
```

> **NOTE:** You can use [Symphony Expressions](../uom/property-expressions.md) in configuration properties

