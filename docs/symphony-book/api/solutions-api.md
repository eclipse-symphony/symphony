# Solutions API
| Route | Method| Function |
|--------|-------|--------|
| ```/solutions/{solution name}``` | POST | Creates or updates a solution |
| ```/solutions/[{solution name}]?[<path=<json path>]&[<doc-type>=<doc type>]``` | GET | Query solutions |
| ```/solutions/{solution name}``` | DELETE | Deletes a solution |

>**NOTE**: ```{}``` indicate path parameter; ```<>``` indicates query parameter; ```[]``` indicates optional parameter

## Query Solutions
* **Path:** /solutions/{solution name}
* **Method:** Get
* **Parameters:**
  |Parameter| Value|
  |--------|--------|
  | ```[{solution name}]``` | (optional) name of the solution, a list is returned when this parameter is omitted |
  | ```[<path>]``` | (option) JSON Path filter|
  |```[<doc-type>]```| (optional) return doc type, like ```yaml``` or ```json```, default is ```json```. See [query projection](./projection.md) for mroe details |
  
* **Headers:**
  |Parameter| Value|
  |--------|--------|
  | ```Authorization``` | Bearer token. see [authorization](../security/authorization.md) for more details |
* **Request Body:** None
* **Response Body** (without [projection](../api/projection.md))**:**
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
  ## Create or Update Solution
* **Path:** /solutions/{target name}
* **Method:** POST
* **Parameters:**
  |Parameter| Value|
  |--------|--------|
  | ```{solution name}``` | name of the solution |
  | ```[<embed-type>]``` | type of embedded artifact<sup>1</sup>|
  | ```[<embed-component>]``` | name of the embedded component<sup>1</sup>|
  | ```[<embed-property>]```| name of the embedded component property<sup>1</sup>|
  
  <sup>1</sup>: Symphony allows a foeign application artifact to be directly embedded as a component property. The embedded artifact can be in any format. When you send the POST request, you need to use the correct content encoding. For example, to embed a YAML file as a ```embedded``` property of a ```my-external-component``` component, you need to set body encoding to ```application/text```.
* **Headers:**
  |Parameter| Value|
  |--------|--------|
  | ```Authorization``` | Bearer token. see [authorization](../security/authorization.md) for more details |
* **Request Body:** Solution Spec (JSON), or an arbitary text payload when ```embed-*``` parameters are used. For example, the following request: ```http://localhost:8080/v1alpha2/solutions/api-test-1?embed-type=container&embed-component=galaxy-services&embed-property=embedded``` with request body:
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
  creates a ```api-test-1``` solution with a single ```galaxy-services``` component that has a ```embedded``` property containing the posted YAML. This is equivalent to sending request to ```http://localhost:8080/v1alpha2/solutions/api-test-1``` with JSON request body:
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
## Delete a Solution
* **Path:** /targets/solutions/{solution name}
* **Method:** DELETE
* **Parameters:**
  |Parameter| Value|
  |--------|--------|
  | ```{solution name}``` | name of the solution |
  
* **Headers:**
  |Parameter| Value|
  |--------|--------|
  | ```Authorization``` | Bearer token. see [authorization](../security/authorization.md) for more details |
* **Request Body:** None
* **Response Body:** None