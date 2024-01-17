# Targets API

| Route | Method| Function |
|--------|-------|--------|
| `/targets/bootstrap` | POST | Bootstraps a target with a target registry. |
| `/targets/download/{doc-type}/{name}?[<path>=<path filter>]` | GET | Target requests downloading artifacts. |
| `/targets/ping/{name}`| GET | Target reports heartbeat signals. |
| `/targets/registery/[{target name}]?[<path=<json path>]&[<doc-type>=<doc type>]`| GET | Get a target. |
| `/targets/status/{name}/{component?}?<status>=<value>` | PUT | Target reports status. |

>**NOTE**: `{}` indicate path parameter; `<>` indicates query parameter; `[]` indicates optional parameter.

## Query targets

* **Path:** /targets/registry/{target name}
* **Method:** Get
* **Parameters:**

  |Parameter| Value|
  |--------|--------|
  | `[{target name}]` | (optional) Name of the target. A list is returned when this parameter is omitted. |
  | `[<path>]` | (option) JSON path filter. |
  |`[<doc-type>]`| (optional) Return doc type, like `yaml` or `json`. Default is `json`. For more information, see [query projection](./projection.md). |
  
* **Headers:**

  |Parameter| Value|
  |--------|--------|
  | `Authorization` | Bearer token. For more information, see [authorization](../security/authorization.md). |

* **Request body:** None
* **Response body** (without [projection](../api/projection.md))**:**

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

## Create or update a target

* **Path:** /targets/registry/{target name}
* **Method:** POST
* **Parameters:**

  |Parameter| Value|
  |--------|--------|
  | `{target name}` | name of the target |
  | `[<with-binding>]` | (option) add a [role binding](../concepts/unified-object-model/target.md#role-bindings). Currently supported binding is `staging`, which binds container operations to a [staging provider](../providers/staging_provider.md)|

* **Headers:**

  |Parameter| Value|
  |--------|--------|
  | `Authorization` | Bearer token. For more information, see [authorization](../security/authorization.md). |

* **Request body:** Target Spec. For example:

  ```json
  {
    "displayName": "my-phone-a",
    "properties": {
      "group": "group-a",
      "foo": "bar"
    }
  }
  ```

  The above query is equivalent to a query without the `with-binding=staging` parameter:

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

* **Response body:** None

## Delete a target

* **Path:** /targets/registry/{target name}
* **Method:** DELETE
* **Parameters:**

  |Parameter| Value|
  |--------|--------|
  | `{target name}` | Name of the target. |
  
* **Headers:**
  |Parameter| Value|
  |--------|--------|
  | `Authorization` | Bearer token. For more information, see [authorization](../security/authorization.md). |
* **Request body:** None
* **Response body:** None

## Target heartbeat

You can optionally send heartbeat signals. When the heartbeat URL is invoked, the current UTC timestamp is saved as a `ping` property in the target status.

* **Path:** /targets/ping/{target name}
* **Method:** GET
* **Parameters:**

  |Parameter| Value|
  |--------|--------|
  | `{target name}` | Name of the target. |
  
* **Headers:**

  |Parameter| Value|
  |--------|--------|
  | `Authorization` | Bearer token. For more information, see [authorization](../security/authorization.md). |

* **Request body:** None
* **Response body:**

  ```json
  {}
  ```

## Target status report

To report target status, you can report a single status value through the query parameter. Or you can PUT a property bag with multiple properties (or use a combination of both).

* **Path:** /targets/status/{target name}
* **Method:** PUT
* **Parameters:**

  |Parameter| Value|
  |--------|--------|
  | `{target name}` | Name of the target. |
  | `[<status>=<value>]` | A key-value pair representing a status value. For example: `http://localhost:8080/v1alpha2/targets/status/my-phone-a?foo=bar` records a property `foo` with value `bar` on target status. |
  
* **Headers:**

  |Parameter| Value|
  |--------|--------|
  | `Authorization` | Bearer token. For more information, see [authorization](../security/authorization.md). |

* **Request body:** Optional. Can be a property bag in the following format:

  ```json
  {
    "status": {
      "properties": {
        "foo": "bar2"
      }
    }
  }
  ```

* **Response body:**

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
