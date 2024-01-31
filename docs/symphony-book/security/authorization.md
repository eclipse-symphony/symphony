# Authorization

Symphony API supports authorization using a [JWT token](https://jwt.io/). To access protected paths, a client needs to attach a Bearer token in its **Authorization** header. Authorization is configured by adding a [JWT handler](../bindings/jwt-handler.md) to the [HTTP binding](../bindings/http-binding.md) [pipeline](../bindings/http-binding.md#pipeline).

## Authorization with tokens

The [Microsoft identity platform](https://learn.microsoft.com/entra/identity-platform/security-tokens) authenticates users and provides security tokens. To use an access token with Symphony, configure your [JWT handler](../bindings/jwt-handler.md) in your HTTP binding pipeline accordingly.

The following example shows that `verifyKey` is set to a certificate public key. It shows requires presence of an `appid` claim and an `oid` claim with specified values.

```json
"pipeline": [
  {
    "type": "middleware.http.jwt",                   
    "properties": {            
      "verifyKey": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAsfsXMXWuO+dniLaIELa3\nPyqz9Y/rWff/AVrCAnFSdPHa8//Pmkbt/yq+6Z3u1o4gjRpKWnrjxIh8zDn1Z1RS\n26nkKcNg5xfWxR2K8CPbSbY8gMrp/4pZn7tgrEmoLMkwfgYaVC+4MiFEo1P2gd9m\nCdgIICaNeYkG1bIPTnaqquTM5KfT971MpuOVOdM1ysiejdcNDvEb7v284PYZkw2i\nmwqiBY3FR0sVG7jgKUotFvhd7TR5WsA20GS/6ZIkUUlLUbG/rXWGl0YjZLS/Uf4q\n8Hbo7u+7MaFn8B69F6YaFdDlXm/A0SpedVFWQFGzMsp43/6vEzjfrFDJVAYkwb6x\nUQIDAQAB\n-----END PUBLIC KEY-----\n",
      "mustHave": ["appid", "oid"],
      "mustMatch": {
        "appid": "<some Microsoft Entra ID app id>",
        "oid": "<some Microsoft Entra ID object id>"
      }
    }
  }
]
```

Use this sample command to generate a Microsoft Entra ID token:

```bash
curl -X POST -H 'Content-Type: application/x-www-form-urlencoded' -d 'grant_type=client_credentials&client_id=<client-id>&resource=2ff814a6-3304-4ab8-85cb-cd0e6f879c1d&client_secret=<application-secret>' https://login.microsoftonline.com/<tenant-id>/oauth2/token
```

## Authorization with the default user store

Symphony offers a default user store that supports basic authentication with a password, mostly for dev-test scenarios. To sign in, send a POST request to `http://<symphony api address>/v1alpha2/users/auth` with the following JSON payload:

```json
{
  "username": "<user name>",
   "password": "<password>"
}
```

A successful login returns a token, which you can use to authorize your Symphony API calls:

```json
{
  "accessToken": "...",
  "tokenType": "Bearer"
}
```

## Role-based access control

Multiple levels of role-based access control (RBAC) can be applied to Symphony:

* Symphony REST API has RBAC control based on claims in authorization tokens
* Kubernetes RBAC can be applied to Kubernetes resources
* When a particular [provider](../providers/providers.md) is used, the provider is subject to authentication and authorization mechanisms reinforced by the connected system.

Kubernetes RBAC and provider configuration are out of scope of this document. This section covers how to configure Symphony REST API RBAC.

### Symphony REST API roles

Symphony REST API defines a few roles:

* **Administrator**: Full access to all Symphony APIs
* **Reader**: - Read-only access to Symphony APIs
* **Solution creator**: CRUD on [solutions](../concepts/unified-object-model/solution.md) only
* **Target manager**: CRUD on [targets](../concepts/unified-object-model/target.md) only
* **Operator**: CRUD on [instances](../concepts/unified-object-model/instance.md) only

Symphony's JWT handler allows you to map claim values into the above roles. For example, the following mapping rule maps an `admin` user to the `administrator` role. When integrated with an external identity provider (IdP) such as Microsoft Entra ID, you probably want to configure such mappings based on the Microsoft Entra ID token you expect (such as looking for embedded application roles).

```json
{
  "role": "administrator",
  "claim": "user",
  "value": "admin"
}
```

A user can be associated with multiple roles. For example, the following mapping assigns any users to the `reader` role. If both mapping rules are applied, the `admin` user is assigned to both the `administrator` role and the `reader` role in this case.

```json
{
  "role": "reader",
  "claim": "user",
  "value": "*"
}
```

### Symphony REST API access policy

You can define REST API path access policies as part of the JWT handler's configuration. For each role, you can define a list of paths and their corresponding HTTP verbs. The following configuration shows a typical configuration of Symphony API:

```json
"pipeline": [
  {
    "type": "middleware.http.jwt",                   
    "properties": {            
      "ignorePaths": ["/v1alpha2/users/auth"],
      "verifyKey": "...",
      "mustHave": ["appid", "oid"],
      "mustMatch": {
        "appid": "<some Microsoft Entra ID app id>",
        "oid": "<some Microsoft Entra ID object id>"
      },
      "roles": [
        {
          "role": "administrator",
          "claim": "user",
          "value": "admin"
        },
        {
          "role": "reader",
          "claim": "user",
          "value": "*"
        },
        {
          "role": "solution-creator",
          "claim": "user",
          "value": "developer"
        },
        {
          "role": "target-manager",
          "claim": "user",
          "value": "device-manager"
        },
        {
          "role": "operator",
          "claim": "user",
          "value": "solution-operator"
        }
      ],
      "policy": {                
        "administrator": {
          "items": {
            "*": "*"                    
          }
        },
        "reader": {
          "items": {
            "*": "GET"
          }
        },
        "solution-creator": {
          "items": {
            "/v1alpha2/solutions": "*"
          }
        },
        "target-manager": {
          "items": {
            "/v1alpha2/targets": "*"
          }
        },
        "solution-operator": {
          "items": {
            "/v1alpha2/instances": "*"
          }
        }                
      }
    }
  }
]
```

## Use an external user store

By default, Symphony uses an in-memory user store to simplify deployments. In a production environment, you'll want to switch to an external user store, such as SQL Server, Redis, or MySQL. Symphony is integrated with [Dapr](https://dapr.io/) through an HTTP state provider accessing the Dapr sidecar state interface. This allows Symphony to connect to a few dozens of database types supported by Dapr.

> **NOTE**: Symphony doesn't write passwords to databases. Instead, it writes a hash based on user id and password. If you plan to support renaming user ids, you need to be aware that the saved hash won't work under the new user id.
