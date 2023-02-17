# JWT Handler

JWT Handler retrieves and verifies a [JWT token](https://jwt.io/) from an authorization header in the request and allows the request to be handled only when the token can be verified.

JWT handler is plugged into a [HTTP binding](../bindings/http-binding.md) via the binding’s [pipeline](../bindings/http-binding.md#pipeline) configuration, for example:
```json
"pipeline": [
    {
        "type": "middleware.http.jwt",                   
        "properties": {
            "ignorePaths": ["/v1alpha2/users/auth", "/v1alpha2/solution/instances"],
            "verifyKey": "SymphonyKey"
        }
    }
]
```
## Handler Configuration
|Property|Value|
|--------|--------|
|```authHeader```|Authorization header name, default is ```Authorization```.|
|```ignorePath```|Paths to be excluded from authorization, as a string array.|
|```verifyKey```|Token verification key<sup>1</sup> |
|```mustHave```|Required claims in the token. Values are not checked, as a string array. To check claim values, use ```mustHave```|
|```mustMatch```|Required claims with specified values<sup>2</sup>.|

<sup>1</sup>: Verification key can be a shared secret or a public key (starts wtih ```-----BEGIN PUBLIC KEY-----```).
<sup>2</sup>: Sample ```mustMatch``` config:
```json
"mustMatch": {
    "subject": "some-subject",
    "foo": "bar",
    "iat": 1516239022.0
}
```