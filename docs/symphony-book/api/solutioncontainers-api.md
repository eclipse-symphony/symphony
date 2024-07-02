# Solution Containers API

| Route | Method| Function |
|--------|-------|--------|
| `/solutioncontainers/{solution container name}` | POST | Creates or updates a solution container |
| `/solutioncontainers/[{solution container name}]?[<path=<json path>]&[<doc-type>=<doc type>]` | GET | Queries solution containers |
| `/solutioncontainers/{solution container name}` | DELETE | Deletes a solution |

>**NOTE**: 
- `{}` indicates a path parameter; `<>` indicates a query parameter; `[]` indicates a optional parameter
- Other container objects `catalog container`, `campaign container` support the same REST APIs.

## Query solution containers

* **Path:** /solutioncontainers/{solution container name}
* **Method:** Get
* **Parameters:**

  |Parameter| Value|
  |--------|--------|
  | `[{solution container name}]` | (optional) Name of the solution container. A list is returned when this parameter is omitted. |
  | `[<path>]` | (optional) JSON path filter. |
  |`[<doc-type>]`| (optional) Return doc type, like `yaml` or `json`. Default is `json`. For more information, see [query projection](./projection.md). |
  
* **Headers:**

  |Parameter| Value|
  |--------|--------|
  | `Authorization` | Bearer token. For more information, see [authorization](../security/authorization.md). |

* **Request body:** None
* **Response body** (without [projection](./projection.md))**:**

  Solution Container GET

  ```json
  {
    "id": "{solution container name}",
    "spec": {...}
  }
  ```

  Solution Containers LIST

  ```json
  [
    {
      "id": "solution container 1",
      "spec": {...}
    },
    ...
    {
      "id": "solution container n",
      "spec": {...}
    },
  ]
  ```

## Create or update a solution container

* **Path:** /solutioncontainers/{solution container name}
* **Method:** POST
* **Parameters:**

  |Parameter| Value|
  |--------|--------|
  | `{solution container name}` | Name of the solution container|

* **Headers:**

  |Parameter| Value|
  |--------|--------|
  | `Authorization` | Bearer token. For more information, see [authorization](../security/authorization.md).  |

* **Request body:** Container container spec. For example:

  ```json
  {
    "name": "samplecontainer",
  }
  ```
* **Response body:** None

## Delete a solution container

* **Path:** /solutioncontainers/{solution container name}
* **Method:** DELETE
* **Parameters:**

  |Parameter| Value|
  |--------|--------|
  | `{solution name}` | Name of the solution |
  
* **Headers:**

  |Parameter| Value|
  |--------|--------|
  | `Authorization` | Bearer token. For more information, see [authorization](../security/authorization.md). |

* **Request body:** None
* **Response body:** None
