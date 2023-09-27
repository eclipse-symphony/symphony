# MQTT Binding

MQTT binding allows you to access Symphony API, standalone providers, as well as Symphony Agent through [MQTT](https://mqtt.org/). Given many enterprises allow MQTT protocol for IoT payloads, the MQTT binding allows Symphony API to be used without extra firewall settings.


## Setting up a MQTT broker
You can use any standard MQTT broker, either cloud-based or locally hosted. This section provides a couple of options using [Eclipse Mosquitto](https://mosquitto.org/).

> **NOTE:** At the time of writing, Symphony MQTT implementation doesn't support authentication, so anonymous access is required.


### Run Eclipse Mosquitto for local tests

1. Create a ```mosquitto.conf``` with the following content:
```bash
persistence false
allow_anonymous true
connection_messages true
log_type all
listener 1883
```
> **NOTE:** You can find a copy of this file under [symphony-docs/samples/mqtt-binding/mosquitto.conf](../../samples/mqtt-binding/mosquitto.conf)

2. Launch Mosquitto with Docker

 ```bash
 docker run --name mosquitto -d -p 1883:1883 --rm -v `pwd`/mosquitto.conf:/mosquitto/config/mosquitto.conf eclipse-mosquitto
 ```

 ### Run Eclipse Mosquitto on K8s
 You can find a sample deployment spec under ```symphony-docs/samples/mqtt-binding/k8s.yaml```. Simply apply it to your kubernetes cluster:
 ```bash
 kubectl create -f ./k8s.yaml
 ```
 ## Configuring MQTT binding

 > **NOTE:** At the time of writing, MQTT binding doesn't support middleware yet.

 To use MQTT binding, define your binding in your [Symphony host configuration file](../hosts/overview.md):
 ```json
  "bindings": [
    {
      "type": "bindings.mqtt",
      "config": {
        "brokerAddress": "tcp://<IP of your MQTT broker>:1883",
        "clientID": "<Client ID of your choice",
        "requestTopic": "coa-request",
        "responseTopic": "coa-response"
      }
    }
  ]
```

Note the topics ```coa-request``` and ```coa-response``` should match with what [MQTT proxy provider](../providers/mqtt_proxy_provider.md) uses when you connect to the proxy provider.
