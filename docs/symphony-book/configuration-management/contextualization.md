# Contextualization and calculated configurations 

Symphony separates application configuration concerns from infrastructural concerns: you model your applications with Solutions and you manage your compute infrastructure with Targets. And with Symphony expressions, you can assemble both relevant infrastructural information and application settings into a context-aware configuration file.
 
This means instead of creating m X n configurations for m applications on n machines, you can create m application configurations and n machine configurations and combine them together on-the-fly. This method reduces duplications and improves consistency. 

For example, your configuration object can refer to an `IP` attribute of the Target to which the application is deployed:
```yaml
${{$property(IP)}}
```
Leveraging Symphony expressions, you can perform certain calculations on configuration values as well. For example, to choose a different Docker image tag based on the Target operation system, you can use a conditional expression that selects a tag postfix based on the `OS` property of a Target:

```yaml
Image-postfix: ${{$if($eq(property(OS), "Windows", ":win', ":linux"))}
```     

