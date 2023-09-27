# Instances API
| Route | Method| Function |
|--------|-------|--------|
| ```/instances/{instance name}``` | POST | Creates or updates an instance |
| ```/instances/[{instance name}]?[<path=<json path>]&[<doc-type>=<doc type>]``` | GET | Query instances |
| ```/instances/{instance name}``` | DELETE | Deletes an instance |

>**NOTE**: ```{}``` indicate path parameter; ```<>``` indicates query parameter; ```[]``` indicates optional parameter

## Query Instances
* **Path:** /instances/{instance name}
* **Method:** Get
* **Parameters:**
  |Parameter| Value|
  |--------|--------|
  | ```[{instance name}]``` | (optional) name of the instance, a list is returned when this parameter is omitted |
  | ```[<path>]``` | (option) JSON Path filter|
  |```[<doc-type>]```| (optional) return doc type, like ```yaml``` or ```json```, default is ```json```. See [query projection](./projection.md) for mroe details |
  
* **Headers:**
  |Parameter| Value|
  |--------|--------|
  | ```Authorization``` | Bearer token. see [authorization](../security/authorization.md) for more details |
* **Request Body:** None
* **Response Body** (without [projection](../api/projection.md))**:**
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
  ## Create or Update Instance
* **Path:** /instances/{instance name}
* **Method:** POST
* **Parameters:**
  |Parameter| Value|
  |--------|--------|
  | ```{instance name}``` | name of the instance |
  | ```[<solution>]``` | (optional) solution to deploy |
  | ```[<target>]```] | (optional) deployment target |
  | ```[<target-selector>]``` | (optional) target selector |
* **Headers:**
  |Parameter| Value|
  |--------|--------|
  | ```Authorization``` | Bearer token. see [authorization](../security/authorization.md) for more details |
* **Request Body:** Instance Spec. For example:
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
  The above query is equivalent to a query with the ```solution=sample-app-v1&target-selector=group%3dgroup-1``` parameter.
* **Response Body:** None

## Delete a Instance
* **Path:** /instances/{instance name}
* **Method:** DELETE
* **Parameters:**
  |Parameter| Value|
  |--------|--------|
  | ```{instance name}``` | name of the instance |
  
* **Headers:**
  |Parameter| Value|
  |--------|--------|
  | ```Authorization``` | Bearer token. see [authorization](../security/authorization.md) for more details |
* **Request Body:** None
* **Response Body:** None