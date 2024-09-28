# Stage Isolation With Provider Proxy

By default, all stage providers are invoked as in-proc calls on the Symphony control plane. This means all stage providers use the same service account context configured for the Symphony API. As the control plane often needs to manage a large number of resources, having a super user with access to all resources is an obvious security concern. A stage provider proxy allows a stage to be processed in an isolated environment such as a separate process, container, virtual machine, or physic device. The stage provider proxy expects a web server (called a **stage runner**) that implements the required Symphony stage provider interface. Although you can use your own stage runner implementations, we recommend using the default Symphony implementation that supports all existing Symphony stage providers to be used over the proxy. 

Such isolation has some distinct benefits:

* Support different execution environments. As a stage runner can be hosted independently from the control plane, the stage runner can be configured with the exact toolchains for the specific stages. For example, a containerized stage runner can have all necessary tools pre-installed. Another example is that a Windows-based stage runner can use Windows toolchains.
* Because a stage runner runs in a different process, you can assign just enough access rights to the process to perform stage activities. 
* The isolation also provides certain protection over vouge provider implementations, such as script-based attacks.
* Resources required by the runner can be mounted locally without needing to be shared with the control plane.

The following diagram illustrates how stage isolation works with the provider proxy and a stage runner:

![stage isolation](../images/stage-isolation.png)

## Decarling stage proxy

You can attach a proxy setting to any of the stage specs. For example, the follow stage spec specifies that the mock stage should be carried out remotely through a processor proxy:

```yaml
stages:
  mock:
    name: mock
    provider: providers.stage.mock
    proxy:
      provider: providers.stage.proxy.http
      config:
        baseUrl: http://localhost:9082/v1alpha2/
        user: admin
        password: ""        
```

## Launch a stage runner using Symphony API binary

You can launch a stage runner by launching the `symphony-api` process with a `symphony-processor-server.json` config:
```bash
./symphony-api -c ./symphony-processor-server.json -l Debug
```