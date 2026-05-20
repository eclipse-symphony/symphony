# SolutionVersion Containers API

| Route | Method| Function |
|--------|-------|--------|
| `/solutions/{solutionversion container name}` | POST | Creates or updates a solutionversion container |
| `/solutions/[{solutionversion container name}]?[<path=<json path>]&[<doc-type>=<doc type>]` | GET | Queries solutionversion containers |
| `/solutions/{solutionversion container name}` | DELETE | Deletes a solutionversion |

>**NOTE**: 
- `{}` indicates a path parameter; `<>` indicates a query parameter; `[]` indicates a optional parameter
- Other container objects `catalogversion container`, `campaignversion container` support the same REST APIs.

## Query solutionversion containers

* **Path:** /solutions/{solutionversion container name}
* **Method:** Get
* **Parameters:**

  |Parameter| Value|
  |--------|--------|
  | `[{solutionversion container name}]` | (optional) Name of the solutionversion container. A list is returned when this parameter is omitted. |
  | `[<path>]` | (optional) JSON path filter. |
  |`[<doc-type>]`| (optional) Return doc type, like `yaml` or `json`. Default is `json`. For more information, see [query projection](./projection.md). |
  
* **Headers:**

  |Parameter| Value|
  |--------|--------|
  | `Authorization` | Bearer token. For more information, see [authorization](../security/authorization.md). |

* **Request body:** None
* **Response body** (without [projection](./projection.md))**:**

  SolutionVersion Container GET

  ```json
  {
    "id": "{solutionversion container name}",
    "spec": {...}
  }
  ```

  SolutionVersion Containers LIST

  ```json
  [
    {
      "id": "solutionversion container 1",
      "spec": {...}
    },
    ...
    {
      "id": "solutionversion container n",
      "spec": {...}
    },
  ]
  ```

## Create or update a solutionversion container

* **Path:** /solutions/{solutionversion container name}
* **Method:** POST
* **Parameters:**

  |Parameter| Value|
  |--------|--------|
  | `{solutionversion container name}` | Name of the solutionversion container|

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

## Delete a solutionversion container

* **Path:** /solutions/{solutionversion container name}
* **Method:** DELETE
* **Parameters:**

  |Parameter| Value|
  |--------|--------|
  | `{solutionversion name}` | Name of the solutionversion |
  
* **Headers:**

  |Parameter| Value|
  |--------|--------|
  | `Authorization` | Bearer token. For more information, see [authorization](../security/authorization.md). |

* **Request body:** None
* **Response body:** None
