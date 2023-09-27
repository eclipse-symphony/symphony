# Defining Configurations

In Symphony, a configuration object is described by a ```Catalog``` object, which can hold a collection of key-value pairs, as shown in the following example:

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

## Static Overrides

You can make a configuration to inherit from another configuration by setting its parentName property to another configuration ```Catalog``` object. The child configuration inherits all values from its parent. And the child can overrides inherited values by redefining these values in its own definition.

## Dynamic Overrides

When you try to resolve a configuration using a ```$config()``` expression, you can specify a list of overrides, such as ```$config(site-config, setting-key, line-config1, line-config2)```. In this case, Symphony will try to resolve the ```setting-key``` value from the ```line-config1``` object and falls back to ```line-config2``` and eventually ```site-config``` if the key is not found.

## Composition
In any of the configuration properties, you can refer to another configuration object, or a specific field of another configuration object. For example, ```$config(<line-tags>, '')``` copies all properties of the ```line-tags``` configuration as sub-properties of the current configuration property. And ```$config(<line-tags>, 'SQL_SERVER')``` copies the ```SQL_SERVER``` property from the ```line-tags``` object.