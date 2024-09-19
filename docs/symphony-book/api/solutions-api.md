# Solutions API

| Route | Method| Function |
|--------|-------|--------|
| `/solutions/{solution name}` | POST | Creates or updates a solution |
| `/solutions/[{solution name}]?[<path=<json path>]&[<doc-type>=<doc type>]` | GET | Query solutions |
| `/solutions/{solution name}` | DELETE | Deletes a solution |

>**NOTE**: `{}` indicate path parameter; `<>` indicates query parameter; `[]` indicates optional parameter

## Query solutions

* **Path:** /solutions/{solution name}
* **Method:** Get
* **Parameters:**

  |Parameter| Value|
  |--------|--------|
  | `[{solution name}]` | (optional) Name of the solution. A list is returned when this parameter is omitted. |
  | `[<path>]` | (option) JSON path filter. |
  |`[<doc-type>]`| (optional) Return doc type, like `yaml` or `json`. Default is `json`. For more information, see [query projection](./projection.md). |
  
* **Headers:**

  |Parameter| Value|
  |--------|--------|
  | `Authorization` | Bearer token. For more information, see [authorization](../security/authorization.md). |

* **Request body:** None
* **Response body** (without [projection](./projection.md))**:**

  Single solution

  ```json
  {
    "id": "{solution name}",    
    "spec": {...}  //solution spec
  }
  ```

  Solution list

  ```json

  [
    {
      "id": "solution name 1",
      "spec": {...}  //solution spec
    },
    ...
    {
      "id": "solution name n",
      "spec": {...}  //solution spec
    },
  ]
  ```

## Create or update a solution

* **Path:** /solutions/{solution name}
* **Method:** POST
* **Parameters:**

  |Parameter| Value|
  |--------|--------|
  | `{solution name}` | Name of the solution. |
  | `[<embed-type>]` | Type of embedded artifact<sup>1</sup>. |
  | `[<embed-component>]` | Name of the embedded component<sup>1</sup>. |
  | `[<embed-property>]`| Name of the embedded component property<sup>1</sup>. |
  
  <sup>1</sup> Symphony allows a foreign application artifact to be directly embedded as a component property. The embedded artifact can be in any format. When you send the POST request, you need to use the correct content encoding. For example, to embed a YAML file as an `embedded` property of a `my-external-component` component, you need to set body encoding to `text/plain`.

* **Headers:**

  |Parameter| Value|
  |--------|--------|
  | `Authorization` | Bearer token. For more information, see [authorization](../security/authorization.md). |

* **Request body:** Solution spec (JSON), or an arbitrary text payload when `embed-*` parameters are used. For example, the following request:

  ```query
  http://localhost:8080/v1alpha2/solutions/api-test-1?embed-type=container&embed-component=galaxy-services&embed-property=embedded
  ```

  with request body:

  ```yaml
  version: '3.8'
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
  ```

  creates an `api-test-1` solution with a single `galaxy-services` component that has an `embedded` property containing the posted YAML. This is equivalent to sending a request to `http://localhost:8080/v1alpha2/solutions/api-test-1` with JSON request body:

  ```json
  {
    "displayName": "api-test-1",
    "components": [
      {
        "name": "galaxy-services",
        "type": "container",
        "properties": {
          "embedded": "version: '3.8'\r\nprovisioner-version: '1.0'\r\nservices:\r\n  my-apache-app:\r\n  download-image: true\r\n  image: docker.io/httpd:2.4\r\n  ports:\r\n  - 8085:80\r\n  volumes:    \r\n  - \"/:/usr/local/apache2/htdocs\"\r\n  container_name: my-apache-app"
        }
      }
    ]
  }
  ```

## Delete a solution

* **Path:** /targets/solutions/{solution name}
* **Method:** DELETE
* **Parameters:**

  |Parameter| Value|
  |--------|--------|
  | `{solution name}` | Name of the solution. |
  
* **Headers:**

  |Parameter| Value|
  |--------|--------|
  | `Authorization` | Bearer token. For more information, see [authorization](../security/authorization.md). |

* **Request body:** None
* **Response body:** None
