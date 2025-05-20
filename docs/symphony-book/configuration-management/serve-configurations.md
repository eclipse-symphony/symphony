# Serving a configuration object

Symphony separates configuration management and configuration serving. Before configurations are served to applications, Symphony performs all necessary resolutions, such as resolving inheritances and filling contextual information.

Configurations can be injected into solution components as environment variables or file mounts (on Kubernetes). An application can also query Symphony REST API to retrieve configurations.

## Configuration in environment variables

A solution component can inject configuration settings as environment variables by defining properties with an `env.` prefix. For example, the following definition injects a `MY_CONFIG` environment variable into the app-container component. The value is a `config-key` field from a `config-obj` catalog.

```yaml
apiVersion: solution.symphony/v1
kind: Solution
metadata: 
  name: csad-featurizer
spec:  
  components:
  - name: app-container
    type: container
    properties:
      container.image: "ghcr.io/eclipse-symphony/sample-flask-app:latest"      
      env.MY_CONFIG: "${{$config(config-obj, config-key)}}" 
```

## Configuration as a mounted file

When deploying applications on Kubernetes, Symphony can assemble a configuration file and mount the file to a pod. To do this, declare a configuration component, and then a container component that has the configuration component as a dependency, as shown in the following example:

```yaml
apiVersion: solution.symphony/v1
kind: Solution
metadata: 
  name: my-app
spec:  
  components:
  - name: app-config
    type: config
    properties:
      appSettings.json: "${{$json($config('<config-obj>', ''))}}"
  - name: app
    type: container
    metadata:
      pod.volumes: "[{\"name\":\"app-config\",\"configMap\":{\"name\":\"app-config\"}}]" 
    properties:
      container.image: "ghcr.io/eclipse-symphony/sample-flask-app:latest"      
      container.volumeMounts: "[{\"name\":\"app-config\",\"mountPath\":\"/app/config\"}]"
    dependencies:
    - app-config
```

What happens behind the scenes during the deployment:

1. Symphony creates a Kubernetes ConfigMap named `app-config`. The ConfigMap contains an `appSettings.json` file, whose content is populated from a `config-obj` catalog object.

2. Symphony creates a pod volume and a volume mount at path `/app/config` for the app container.

## Querying configurations through Symphony REST API

Some applications desire to dynamically retrieve configurations from a preconfigured endpoint when they are launched. Symphony allows this pattern by providing a REST API for configuration queries.

To query a configuration object, you can send an authenticated **GET** request to the /`catalogs/registry` route.

### Query by id

Send a **GET** request to `catalogs/registry/<id>`. You'll get the complete configuration object (as a Catalog object).

> **NOTE:** In this case, the expressions in your configuration object are not evaluated, because Symphony doesn't know the context for evaluation.

### Query by filters

You can also query configurations by filters like labels and specification properties. For example, you can search for a configuration object that is labeled to be used for a specific application context without knowing the exact configuration name. 
For example, the following query returns a configuration labeled for the night shift of an application:
```bash
GET /catalogs/registry/<id>?filterType=label&filterValue=app==my-app,shit==night
```
> **NOTE:** Query parameters are not URL-encoded in the above sample for clarity. Your query parameters should be URL-encoded.

See [state provider](../providers/state-providers/_overview.md) for more details on filters.

### Retrieve specific portions

Symphony query APIs allows object projection using a Json Path syntax. Instead of retrieving the entire catalog object, you can directly retrieve specific portions or specific fields from a configuration object. 
For example, the following query returns only the ```foo``` field of a configuration object:
```bash
GET /catalogs/registry/<id>?doc-type=yaml&path=$.spec.properties.foo
```
See [projection](../api/projection.md) for more details on query projection.

## Configuration resolution

Symphony uses Configuration Providers to resolve configurations. You can configure multiple configuration providers with precedence in Symphony. When trying to resolve a configuration key, Symphony tries all configuration providers until it finds a matching key.

The default configuration provider shipped with Symphony is based on Catalogs objects. However, other providers can be added to resolve configurations from other sources and to provide additional functionality. For example, a caching provider can be added at the top of the precedence list to provide faster lookups, so the configurations don’t need to be repeatedly resolved. 

