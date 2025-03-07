# Conditional Activation

## Use Case

There can be sections of configuration that need to be toggled on or off based on variables describing the object.  

### Example

#### Prior to Symphony configuration management

A combination of templates and variables could be used along with a script to compile the final configuration object.

`IS_PREMIUM_LINE` will be used in the template to toggle a section on if true, or off if false.

```json
"lines": {
    "line1": {
        "variables": {
            "IS_PREMIUM_LINE": true
        }
    },
    "line2": {
        "variables": {
            "IS_PREMIUM_LINE": false
        }
    }
}
```

This is an example template that utilizes that variable.  This block is included in the final configuration object only if the variable evaluates to true within the context it is applied.

```json
"properties": {
    "premium_properties": {
        "INCLUDE_CHECK": "IS_PREMIUM_LINE",
        "premium_key1": "premium1",
        "premium_key2": "premium2",
    }
}
```

#### Solving via Symphony configuration management

[Symphony property expressions](../../../symphony-book/concepts/unified-object-model/property-expressions.md#functions) allow for conditional inclusion.  

1. Define variables and values per line in a configuration object (catalog).  See [line1](./catalogs/line1.yml) where `IS_PREMIUM_LINE` is true and [line2](./catalogs/line2.yml) where `IS_PREMIUM_LINE` is false.  
1. Define settings to be included based on those variables.  This sample has some settings that are only included on premium lines where `IS_PREMIUM_LINE` is true.  See [premium-properties](./catalogs/premium-properties.yml).
1. In the final configuration object, use the `$if()` function described in the property expressions documentation to conditionally include the `premium` configuration object.  See [line-config](./catalogs/line-config.yml).

    ```yml
    spec:
    catalogType: config
    properties:
        EXTRA-LINE: # Added due to config behavior described in bug: https://github.com/eclipse-symphony/symphony/issues/202
            line1: ${{$if($config('line1', 'IS_PREMIUM_LINE'), $config('premium',''), '')}}
            line2: ${{$if($config('line2', 'IS_PREMIUM_LINE'), $config('premium',''), '')}}
    ```

1. Deploy the example from the conditional-activation directory:

    `kubectl apply -f ./ --recursive`
1. Once the instance has reconciled successfully, view the config map:

    `kubectl describe configmap conditional-activation-config`
1. The resulting config map should have premium settings for line 1 only:

    ```json
    "lines":
    {
        {
            "line1": {
                "premium_key1": "premium1",
                "premium_key2": "premium2"
            },
            "line2": ""
        }
    }
    ```
