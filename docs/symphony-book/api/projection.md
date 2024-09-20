# Query projection

Many Symphony query API interfaces allow you to project query data using a JSONpath expression. This allows you to transform the default data schema to meet with your expectations. Generally, projects are controlled by two parameters, `path` and `doc-type`. The `path` parameter specifies the JSONpath you want to use, and the `doc-type` parameter specifies the result encoding, which can be either json or yaml.

## Example: List target ids

The following query returns a JSON array of target ids:

```query
http://localhost:8080/v1alpha2/targets/registry?path=$.id
```

Sample result:

```json
[
  "my-phone-1",
  "my-phone-2"
]
```

## Example: List target display names

The following query returns a JSON array that contains target names:

```query
http:///localhost:8080/v1alpha2/targets/registry/?path=$.spec.displayName
```

Sample result:

```json
[
  "my-phone-1",
  "my-phone-2"
]
```

## Example: List target specs

The following query returns a list of target specs:

```query
http://localhost:8080/v1alpha2/targets/registry?path=$.spec
```

Sample result:

```json
[
  {
    "components": []
    ...
  },
  ...
  {
    "components": []
    ...
  },
]
```

## Example: Extract embedded YAML

Symphony allows artifacts to be directly embedded as object properties. For example, the following target has an YAML file embedded in its `embedded` property:

```yaml
apiVersion: fabric.symphony/v1
kind: Target
metadata:
  name: my-phone-2  
spec:
  components:
  - name: galaxy-services
    properties:
      embedded: |
        version: '3.7'
        provisioner-version: '1.0'
        services:
          my-apache-app:
            download-image: true
            image: docker.io/httpd:2.4
            ports:
            - 8085:80
            volumes:
            - "/:/usr/local/apache2/htdocs"
            container_name: my-apache-app
    type: container
  displayName: my-phone-2  
```

To retrieve the embedded YAML, use the following query. Note that when `doc-type=yaml` is used, the returned document is encoded as `text/plain` instead of `application/json`.

```query
http:///localhost:8080/v1alpha2/targets/registry/my-phone-2?path=$.spec.components[0].properties.embedded&doc-type=yaml
```

The above query returns a YAML file:

```yaml
provisioner-version: "1.0"
services:
  my-apache-app:
    container_name: my-apache-app
    download-image: true
    image: docker.io/httpd:2.4
    ports:
    - 8085:80
    volumes:
    - /:/usr/local/apache2/htdocs
version: "3.7"
```

> **NOTE**: you can use a shorthand name `first_embedded`instead of `.spec.components[0].properties.embedded`.
