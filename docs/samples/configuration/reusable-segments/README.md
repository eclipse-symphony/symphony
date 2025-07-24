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
        catalogType: config
        properties:
            common-key1: "common-value1"
            common-key2: "common-value2"
    ```

1. Define a top level configuration object that has its own properties and also uses the shared segment.  See [top-level](./catalogs/top-level.yml).

    ```yml
    spec:
        catalogType: config
        properties:
            reusable-segment: ${{$config('segment', '')}}
            key1: "value1"
            key2: "value2"
    ```

1. Deploy the example from the reusable-segments directory:

    `kubectl apply -f ./ --recursive`
1. Once the instance has reconciled successfully, view the config map:

    `kubectl describe configmap reusable-segment-config`
1. The resulting config map contains the values from the shared segment and the additional values from the top level:

    ```json
    {
        // From shared segment
        "common-key1": "common-value1",
        "common-key2": "common-value2",
        // From top level config
        "key1": "value1",
        "key2": "value2"
    }
    ```
