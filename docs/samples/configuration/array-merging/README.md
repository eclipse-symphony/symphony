# Array Merging

## Use Case

A customer might have a sizable configuration object that significantly overlaps with other large configuration objects. They would benefit from creating a single object containing the shared configuration and then combining it with any unique properties.

### Example

As a simple example of this problem, there are two regions which each have a "Tags" array.  They have an overlap of three tags and each region has additional unique tags.

```json
{
    "Name": "app-rules-region1",
    "Tags": [
        "Tag1",
        "Tag2",
        "Tag3",
        "Tag4",
        "Tag5"
    ]
}
```

```json
{
    "Name": "app-rules-region2",
    "Tags": [
        "Tag1",
        "Tag2",
        "Tag3",
        "Tag6",
        "Tag7",
        "Tag8"
    ]
}
```

#### Solving with Symphony configuration management

[Symphony property expressions](../../../symphony-book/concepts/unified-object-model/property-expressions.md#functions) enable the assembly of a configuration object (catalog) by integrating multiple distinct objects.

1. Break the shared tags out into a segment to be shared.  See [shared-tags](./catalogs/shared-tags.yml).

    ```yml
    spec:
    catalogType: config
    properties:
        tags: [
        "Tag1",
        "Tag2",
        "Tag3"
        ]
    ```

1. Define site/region specific additional tags.  See [region1-tags](./catalogs/region1-tags.yml) and [region2-tags](./catalogs/region2-tags.yml).

    Region 1 Example:

    ```yml
        spec:
        catalogType: config
        properties:
            tags: [
            "Tag4",
            "Tag5"
            ]
    ```

1. Combine the two arrays with the following expression.  See [region1-tags](./catalogs/region1-tags.yml) and [region2-tags](./catalogs/region2-tags.yml).
.

    ```yml
        spec:
        catalogType: config
        properties:
            name: "tags-region1"
            # This will combine the two arrays of strings into one.  The double dollar sign on the second config object is required
            tags: ${{$config('shared:tags', 'tags') $$config('region1:tags', 'tags')}}
    ```

1. Deploy the example from the array-merging directory:

    `kubectl apply -f ./ --recursive`
1. Once the instance has reconciled successfully, view the config map:

    `kubectl describe configmap merged-config-region1`

    `kubectl describe configmap merged-config-region2`
1. The resulting config map for each region should have a tags array that contains the shared segment tags and the region specific tags.

    Region 1

    ```json
        {
            "name": "tags-region1",
            "tags": [
                "Tag1",
                "Tag2",
                "Tag3",
                "Tag4",
                "Tag5"
            ]
        }
    ```

    Region 2

    ```json
        {
            "name": "tags-region2",
            "tags": [
                "Tag1",
                "Tag2",
                "Tag3",
                "Tag6",
                "Tag7",
                "Tag8"
            ]
        }
    ```
