# Context based configuration

## Use Case

The required configuration may vary based on the context of the instance.

### Example

There are two different line configurations, the scope of the instance determines which will be used in the generated config map.

[Symphony property expressions](../../../symphony-book/concepts/unified-object-model/property-expressions.md#functions) allow for retrieval of the evaluation context.  

1. Define the configurations, using a name that will be part of the instance scope.  See [line1](./catalogs/line1.yml) and [line2](./catalogs/line2.yml).

1. Set a scope on the instances.  See [instance-line1](./instances/instance-line1.yml) and [instance-line2](./instances/instance-line2.yml).

1. In the solution, when you are setting the configuration property that needs to vary, utilize the `$context()` property expression to build the name of the config object.

    ```yml
    - name: context-based-config
      type: config
      properties: 
        # This uses the context of the instance's scope to determine which config to include
        app-settings: ${{$config($context('$.Instance.Spec.Scope')-config, '')}}
    ```

1. Deploy the example from the context-based directory:

    `kubectl apply -f ./ --recursive`
1. Once the instances have reconciled successfully, view the config maps:

    `kubectl describe configmap context-based-config -n line1`

    `kubectl describe configmap context-based-config -n line2`
1. The resulting config map has the appropriate configuration in each namespace:

    line1

    ```json
    {
        "capacity": 100,
        "name": "Line 1"
    }
    ```

    line2

    ```json
    {
        "capacity": 200,
        "name": "Line 2"
    }
    ```
