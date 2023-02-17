# Targets API

| Route | Method| Function |
|--------|-------|--------|
| ```/targets/bootstrap``` | POST | Bootstraps a target with a target registry |
| ```/targets/download/{doc-type}/{name}?[<path>=<path filter>]``` | GET | Target requests downloading artifacts |
| ```/targets/ping/{name}```| GET | Target reports heartbeat signals |
| ```/targets/registery/[{target name}]?[<path=<json path>]&[<doc-type>=<doc type>]```| GET | Get an target |
| ```/targets/status/{name}/{component?}?<status>=<value>``` | PUT | Target reports status |

>**NOTE**: ```{}``` indicate path parameter; ```<>``` indicates query parameter; ```[]``` indicates optional parameter

## Query Targets
* **Path:** /targets/registry/{target name}
* **Method:** Get
* **Parameters:**
  |Parameter| Value|
  |--------|--------|
  | ```[{target name}]``` | (optional) name of the target, a list is returned when this parameter is omitted |
  | ```[<path>]``` | (option) JSON Path filter|
  |```[<doc-type>]```| (optional) return doc type, like ```yaml``` or ```json```, default is ```json```. See [query projection](./projection.md) for mroe details |
  
* **Headers:**
  |Parameter| Value|
  |--------|--------|
  | ```Authorization``` | Bearer token. see [authorization](../security/authorization.md) for more details |
* **Request Body:** None
* **Response Body** (without [projection](../api/projection.md))**:**
  Single target
  ```json
  {
    "id": "{target name}",
    "status": {...}, //target status
    "spec": {...}  //target spec
  }
  ```
  Target list
  ```json
  [
    {
        "id": "target name 1",
        "status": {...}, //target status
        "spec": {...}  //target spec
    },
    ...
    {
        "id": "target name n",
        "status": {...}, //target status
        "spec": {...}  //target spec
    },
  ]
  ```

## Create or Update Target
* **Path:** /targets/registry/{target name}
* **Method:** POST
* **Parameters:**
  |Parameter| Value|
  |--------|--------|
  | ```{target name}``` | name of the target |
  | ```[<with-binding>]``` | (option) add a [role binding](../target-management/target-management.md#role-bindings). Currently supported binding is ```staging```, which binds container operations to a [staging provider](../providers/staging_provider.md)|
* **Headers:**
  |Parameter| Value|
  |--------|--------|
  | ```Authorization``` | Bearer token. see [authorization](../security/authorization.md) for more details |
* **Request Body:** Target Spec. For example:
  ```json
  {
    "displayName": "my-phone-a",
    "properties": {
        "group": "group-a",
        "foo": "bar"
    }
  }
  ```
  The above query is equivalent to a query without the ```with-binding=staging``` parameter:
  ```json
  {
    "displayName": "my-phone-a",
    "properties": {
        "group": "group-a",
        "foo": "bar"
    },
    "topologies": [
        {
            "bindings": [
                {
                    "role": "instance",
                    "provider": "providers.target.staging",
                    "config": {
                        "inCluster": "true",
                        "targetName": "my-phone-a"
                    }
                }
            ]
        }
    ]
  }
  ```
* **Response Body:** None

## Delete a Target
* **Path:** /targets/registry/{target name}
* **Method:** DELETE
* **Parameters:**
  |Parameter| Value|
  |--------|--------|
  | ```{target name}``` | name of the target |
  
* **Headers:**
  |Parameter| Value|
  |--------|--------|
  | ```Authorization``` | Bearer token. see [authorization](../security/authorization.md) for more details |
* **Request Body:** None
* **Response Body:** None

## Target Heartbeat
You can optionally send heartbeat signals. When the heartbeat URL is invoked, the current UTC timestamp is saved as a ```ping``` property in the Target status.

* **Path:** /targets/ping/{target name}
* **Method:** GET
* **Parameters:**
  |Parameter| Value|
  |--------|--------|
  | ```{target name}``` | name of the target |
  
* **Headers:**
  |Parameter| Value|
  |--------|--------|
  | ```Authorization``` | Bearer token. see [authorization](../security/authorization.md) for more details |
* **Request Body:** None
* **Response Body:** 
  ```json
  {}
  ```

## Target Status Report
To report Target status, you can report a single status value through the query parameter. Or, you can PUT a property bag with multiple properties (or use a combination of both).
* **Path:** /targets/status/{target name}
* **Method:** PUT
* **Parameters:**
  |Parameter| Value|
  |--------|--------|
  | ```{target name}``` | name of the target |
  | ```[<status>=<value>]``` | a key-value pair representing a status value, for example: ```http://localhost:8080/v1alpha2/targets/status/my-phone-a?foo=bar``` records a property ```foo``` with value ```bar``` on Target status |
  
* **Headers:**
  |Parameter| Value|
  |--------|--------|
  | ```Authorization``` | Bearer token. see [authorization](../security/authorization.md) for more details |
* **Request Body:** Optional. Can be a property bag in the following format:
  ```json
  {
    "status": {
        "properties": {
            "foo": "bar2"
        }
    }
  }
  ```
* **Response Body:** 
  Current Target status
  ```json
  {
    "id": "my-phone-a",
    "status": {
        "deployed": "0",
        "foo": "bar",
        "ping": "2022-10-27 01:16:06.867714383 +0000 UTC",
        "status": "OK",
        "targets": "1",
        "targets.my-phone-a": "OK - "
    }
  }
  ```