# Reusable Segments

## Use Case

There can be sections of configuration that are reused across many objects. Rather than repeating them in every context they are needed, they can be broken out into a separate segment.

### Example

In this small example, there are two common keys that will be reused in multiple objects.  These common keys will be decomposed into a reusable segment.

#### Solving via Symphony configuration management

Composition allows for a configuration object to be made of many other objects.

1. Define shared keys and values, see [segment](./catalogs/segment.yml).

    ```yml
    spec:
        type: config
        properties:
            EXTRA-LINE: # Added due to config behavior described in bug: https://github.com/eclipse-symphony/symphony/issues/202
                common-key1: "common-value1"
                common-key2: "common-value2"
    ```

1. Define a top level configuration object that has its own properties and also uses the shared segment.  See [top-level](./catalogs/top-level.yml).

    ```yml
    spec:
        type: config
        properties:
            reusable-segment: ${{$config('segment', '')}}
            EXTRA-LINE: # Added due to config behavior described in bug: https://github.com/eclipse-symphony/symphony/issues/202
                key1: "value1"
                key2: "value2"
    ```

1. Deploy the example from the reusable-segments directory:

    `kubectl apply -f ./ --recursive`
1. Once the instance has reconciled successfully, view the config map:

    `kubectl describe configmap reusable-segment-config`
1. The resulting config map should have premium settings for line 1 only:

    ```json
    {
        "common-key1": "common-value1",
        "common-key2": "common-value2",
        "key1": "value1",
        "key2": "value2"
    }
    ```
