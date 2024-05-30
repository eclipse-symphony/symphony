# Late Assembly

Symphony breaks away from the file-based mindset. Operations can model and manage configurations in flexible ways using Symphony configuration modeling capabilities. The final configuration that is to be served to the application is assembled at the last moment to ensure the most fresh contextual information is injected.

You can refer to fields in multiple configuration objects in your artifacts, for example:

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
      env.MY_CONFIG: "${{$config(config-a, config-key)}}"
      env.OTHER_CONFIG: "${{$config(config-b, other-key)}}"
```

You can also use recursive expressions to resolve for complex configuration scenarios. For example:

```yaml
${{$config(line-config, $config(scenario-config, active-scenario))}}
```

Such late assembly can happen directly in any artifact formats (such as Solutions). So, you don't have to explicitly define an unified configuration (Catalog) object. On the other hand, if you do want the combined configuration to be explicitly managed (such as to be versioned independently from the application), you can always create a combined configuration object that assembles other configuration objects, or parts of other configuration objects into an unified object:

```yaml
apiVersion: federation.symphony/v1
kind: Catalog
metadata:
  name: robot-config
spec:  
  siteId: hq
  type: config
  properties:
    some-object: "${{$json($config('<config-obj>', ''))}}"
    some-value: "${{$config('other-config','some-field)}}"
```

Symphony allows multiple levels of compositions. If the composed configurations have additional references to yet other configuration objects, the entire reference tree will be resolved

