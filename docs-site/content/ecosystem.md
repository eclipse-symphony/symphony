---
type: docs
title: "Embracing the Ecosystem"
linkTitle: "Ecosystem"
description: ""
weight: 70
---

As a platform-neutral orchestrator, Symphony is designed to work with existing ecosystems through its extensibility model. The following summary highlights how Symphony natively integrates with various projects, services, platforms, and tools, while allowing additional extensions to be introduced at any time.

## Target Providers
Symphony Target Providers are responsible for applying updates to target devices or clusters. Symphony currently includes the following built-in target providers:

* [Azure Device Update for IoT Hub](https://learn.microsoft.com/en-us/azure/iot-hub-device-update/)
* [Azure Resource Manager (ARM)](https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/overview)
* [Bash Script](https://www.w3schools.com/bash/bash_script.php)
* [Docker](https://www.docker.com/)
* [Eclipse Ankaios](https://projects.eclipse.org/projects/automotive.ankaios)
* [Eclipse uProtocol](https://uprotocol.org/)
* [Helm](https://helm.sh/)
* [Kubernetes ConfigMap](https://kubernetes.io/)
* [Kubernetes Ingress](https://kubernetes.io/)
* [Kubernetes Pods](https://kubernetes.io/)
* [PowerShell](https://learn.microsoft.com/en-us/powershell/)
* [RESTful API](https://en.wikipedia.org/wiki/Overview_of_RESTful_API_Description_Languages)
* [Windows 10/11 Sideloading](https://learn.microsoft.com/en-us/windows/application-management/sideload-apps-in-windows)

## Symphony Components
Symphony system components can be swapped out to support different deployment needs and scenarios. Symphony currently ships with the following swappable system components:

* **Identity Providers**
    * Local user/password
    * [Microsoft Azure Atteststation](https://learn.microsoft.com/en-us/azure/attestation/overview)
    * [Microsoft Entra](https://learn.microsoft.com/en-us/entra/fundamentals/what-is-entra) (including managed identities)
    * [OpenID Connect (OIDC)](https://en.wikipedia.org/wiki/OpenID) compatible identity providers

* **Message Bus**
    * In-memory pub-sub
    * [MQTT](https://mqtt.org/) brokers

* **Secret Store**
    * [Kubernetes secrets](https://kubernetes.io/)

* **State Store**
    * In-memroy store
    * [Kubernetes state store](https://kubernetes.io/)
    * [Redis](https://redis.io/)
    