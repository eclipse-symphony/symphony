# Sideload Apps Provider for Windows 10 and XBOX

This provider installs Windows apps through [sideloading](https://learn.microsoft.com/en-us/windows/application-management/sideload-apps-in-windows-10). 

## Component Property Mappings

This provider installs your application package (```.appxbundle```), registered as a **ComponentSpec**, to a target machine. Please see the [target configuration](#target-configuration) section for more details.

**ComponentSpec** Properties are mapped as the following:

| ComponentSpec Properties | Win 10 Sideload Provider |
|--------|--------|
| ```name```| This should match with your application package name|
| ```app.package.path```| Path to the ```.appxbundle``` file. This path should be accessible from the provider, such as a Win 10 sideload provider hosted on a [Windows-based Symphony agent](../build_deployment/windows_agent.md).|

## Provider Configuration

| Field | Comment |
|--------|--------|
| IPAddress | IP address of the target Windows 10 machine |
| Pin | Pairing pin<sup>2</sup>|
| WinAppDeployCmdPath | Full path to ```WinAppDeployCmd.exe```<sup>1</sup>|

<sup>1</sup>: You can install ```WinAppDeployCmd.exe``` through [Windows 10 SDK](https://developer.microsoft.com/en-us/windows/downloads/windows-sdk/). You need at least version 1803.
<sup>2</sup>: Although in theory you can put the pairing pin (see [target configuration](#target-configuration)) in your provider configuration, since this is a one-time pin, you probably want to go through the pairing process once and omit this setting in your configuration.

## Target Configuration
Before you can use sideloading to install apps on your Windows 10 client devices, you need to configure your target devices to:

* Enable the [Developer Mode](https://learn.microsoft.com/en-us/windows/apps/get-started/enable-your-device-for-development).
* When developer mode is enabled, also enable the **Device Discovery** feature. 

### 1. Pairing your agent machine and your target machine
1. On your target machine, under the **Device Discovery** developer feature, click on the **Pair** button to display the paring pin.
2. From the machine where you plan to run your [Windows-based Symphony agent](../build_deployment/windows_agent.md), use command:
    ```cmd
    WinAppDeployCmd list -ip <target machine IP> -pin <pairing pin>
    ```
    This command lists all installed application packages on your target machine. This also remembers the pairing pin so that you don't need the pairing pin any more.
### 2. Configuring the Windows-based Sysmphony agent

Follow instructions here to conigure your agent: [Windows-based Symphony agent](../build_deployment/windows_agent.md). 

### 3. Import application signing certificate
You also need to import the application signing certificate. Please contact your application vendor.

## Create Symphony Target
Because Symphony control plane usually runs on a Linux-based Kubernetes cluster, it uses a [proxy provider](../providers/proxy_provider.md) to talk to the Windows-based Symphony agent you just configured.