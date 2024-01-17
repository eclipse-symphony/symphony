# Serve configurations

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
      container.image: "ghcr.io/azure/symphony/sample-flask-app:latest"      
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
      container.image: "ghcr.io/azure/symphony/sample-flask-app:latest"      
      container.volumeMounts: "[{\"name\":\"app-config\",\"mountPath\":\"/app/config\"}]"
    dependencies:
    - app-config
```

What happens behind the scenes during the deployment:

1. Symphony creates a Kubernetes ConfigMap named `app-config`. The ConfigMap contains an `appSettings.json` file, whose content is populated from a `config-obj` catalog object.

2. Symphony creates a pod volume and a volume mount at path `/app/config` for the app container.
