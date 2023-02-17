# providers.target.azure.iotedge
This provider maps components to Aure IoT Edge module definitions. Symphony allows multiple solutions, as well as multiple solution instances, to be deployed on the same IoT Edge device. It does this by prefixing module names with an instance prefix. And it also automatically rewrites any route definitions so that routes are resolved correctly within an instance. 

## Component Property Mappings

IoT Edge Provider maps a **ComponentSpec** to a ```Module``` and a ```ModuleTwin``` (referred as ```Module``` and ```ModuleTwin``` in the following tables).

| Symphony Object | IoT Edge Provider |
|--------|--------|
|```ComponentSpec.Name```|```ModuleTwin.ModuleId```| 
|```Instance.Metadata["deployment.imagePullSecrets"]```|---|
|```Instance.Metadata["deployment.nodeSelector"]```|---|
|```Instance.Metadata["deployment.replicas"]```|---|
|```Instance.Metadata["deployment.scope"]```|---|
|```Instance.Metadata["deployment.volumes"]```|---|
|```Instance.Metadata["service.annotation.<label>]```|---|
|```Instance.Metadata["service.loadBalancerIP]```|---|
|```Instance.Metadata["service.ports]```|---|
|```Instance.Metadata["service.type]```|---|

**ComponentSpec** Properties are mapped as the following:

| ComponentSpec Properties | IoT Edge Provider |
|--------|--------|
|```container.args```|---|
|```container.commands```|---|
|```container.createOptions```|```Module.Settings.createOptions```|
|```container.image```|```Module.Settings.image```|
|```container.imagePullPolicy```|---|
|```container.ports```|---|
|```container.resources```|---|
|```container.restartPolicy```|```Module.RestartPolicy```|
|```container.type```|```Module.Type```|
|```container.version```|```Module.Version```|
|```container.volumeMounts```|---|
|```desired.<property>```|```ModuleTwin.Properties.Desired.<property>```<sup>(1)</sup> |


(1): Only setting names starting with a letter (a-z or A-Z) are mapped. 