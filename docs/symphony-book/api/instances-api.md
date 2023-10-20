# Instances API

| Route | Method| Function |
|--------|-------|--------|
| `/instances/{instance name}` | POST | Creates or updates an instance |
| `/instances/[{instance name}]?[<path=<json path>]&[<doc-type>=<doc type>]` | GET | Queries instances |
| `/instances/{instance name}` | DELETE | Deletes an instance |

>**NOTE**: `{}` indicates a path parameter; `<>` indicates a query parameter; `[]` indicates an optional parameter

## Query instances

* **Path:** /instances/{instance name}
* **Method:** Get
* **Parameters:**

  |Parameter| Value|
  |--------|--------|
  | `[{instance name}]` | (optional) Name of the instance. A list is returned when this parameter is omitted. |
  | `[<path>]` | (optional) JSON path filter. |
  |`[<doc-type>]`| (optional) Return doc type, like `yaml` or `json`. Default is `json`. For more information, see [query projection](./projection.md). |
  
* **Headers:**

  |Parameter| Value|
  |--------|--------|
  | `Authorization` | Bearer token. For more information, see [authorization](../security/authorization.md). |

* **Request body:** None
* **Response body** (without [projection](./projection.md))**:**

  Single instance

  ```json
  {
    "id": "{instance name}",
    "status": {...}, //instance status
    "spec": {...}  //instance spec
  }
  ```

  Instance list

  ```json
  [
    {
      "id": "instance name 1",
      "status": {...}, //instance status
      "spec": {...}  //instance spec
    },
    ...
    {
      "id": "instance name n",
      "status": {...}, //instance status
      "spec": {...}  //instance spec
    },
  ]
  ```

## Create or update an instance

* **Path:** /instances/{instance name}
* **Method:** POST
* **Parameters:**

  |Parameter| Value|
  |--------|--------|
  | `{instance name}` | Name of the instance |
  | `[<solution>]` | (optional) Solution to deploy |
  | `[<target>]` | (optional) Deployment target |
  | `[<target-selector>]` | (optional) Target selector |

* **Headers:**

  |Parameter| Value|
  |--------|--------|
  | `Authorization` | Bearer token. For more information, see [authorization](../security/authorization.md).  |

* **Request body:** Instance spec. For example:

  ```json
  {
    "name": "sample-app",
    "displayName": "sample-app",
    "solution": "sample-app-v1",
    "target": {
      "selector": {
        "group": "group-1"
      }
    }
  }
  ```

  The above query is equivalent to a query with the `solution=sample-app-v1&target-selector=group%3dgroup-1` parameter.

* **Response body:** None

## Delete an instance

* **Path:** /instances/{instance name}
* **Method:** DELETE
* **Parameters:**

  |Parameter| Value|
  |--------|--------|
  | `{instance name}` | Name of the instance |
  
* **Headers:**

  |Parameter| Value|
  |--------|--------|
  | `Authorization` | Bearer token. For more information, see [authorization](../security/authorization.md). |

* **Request body:** None
* **Response body:** None
