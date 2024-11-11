#
# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
# SPDX-License-Identifier: MIT
#
from dataclasses import dataclass
from typing import List
from typing import Dict
from flask import Flask, abort, request, Response
import jsons
import json
from waitress import serve
import paho.mqtt.client as mqtt

@dataclass
class ObjectMeta:
    namespace: str = ""
    name:  str = "" 
    labels: Dict[str, str] = None
    annotations: Dict[str, str] = None

@dataclass
class TargetSelector:
    name: str = ""
    selector: Dict[str, str] = None

@dataclass
class BindingSpec:
    role: str = ""
    provider: str = ""
    config: Dict[str, str] = None

@dataclass
class TopologySpec:
    device: str = ""
    selector: Dict[str, str] = ""
    bindings: List[BindingSpec] = None

@dataclass
class PipelineSpec:
    name: str = ""
    skill: str = ""
    parameters: Dict[str, str] = None

@dataclass
class VersionSpec:
    solution: str = ""
    percentage: int = 100

@dataclass
class InstanceSpec:
    name: str = ""
    parameters: Dict[str, str] = None
    solution: str = ""
    target: TargetSelector = None
    topologies: List[TopologySpec] = None
    pipelines: List[PipelineSpec] = None
    scope: str = ""    
    displayName: str = ""
    metadata: Dict[str, str] = None
    versions: List[VersionSpec] = None
    arguments: Dict[str, Dict[str, str]] = None
    optOutReconciliation: bool = False
    

@dataclass
class FilterSpec:
    direction: str = ""
    parameters: Dict[str, str] = None
    type: str = ""

@dataclass
class RouteSpec:
    route: str = ""
    properties: Dict[str, str] = None
    filters: List[FilterSpec] = None
    type: str = ""

@dataclass
class ComponentSpec:
    name: str = ""   
    type: str = ""
    routes: List[RouteSpec] = None
    constraints: str = ""
    properties: Dict[str, str] = None
    depedencies: List[str] = None
    skills: List[str] = None
    metadata: Dict[str, str] = None
    parameters: Dict[str, str] = None
    

@dataclass
class SolutionSpec:    
    components: List[ComponentSpec] = None    
    scope: str = ""    
    displayName: str = ""
    metadata: Dict[str,str] = None
    
@dataclass
class SolutionState:
    metadata: ObjectMeta = None
    spec: SolutionSpec = None

@dataclass
class TargetSpec:
    properties: Dict[str, str] = None
    components: List[ComponentSpec] = None
    constraints: str = ""
    topologies: List[TopologySpec] = None
    scope: str = ""
    displayName: str = ""
    metadata: Dict[str, str] = None
    forceRedeploy: bool = False

@dataclass
class ComponentError:
    code: str = ""
    message: str = ""
    target: str = ""    

@dataclass
class TargetError:
    code: str = ""
    message: str = ""
    target: str = ""
    details: Dict[str, ComponentError] = None

@dataclass
class ErrorType:
    code: str = ""
    message: str = ""
    target: str = ""
    details: Dict[str, TargetError] = None

@dataclass
class ProvisioningStatus:
    operationId: str = ""
    status: str = ""
    failureCause: str = ""
    logErrors: bool = False
    error: ErrorType = None
    output: Dict[str, str] = None

@dataclass
class TargetStatus:
    properties: Dict[str, str] = None
    provisioningStatus: ProvisioningStatus = None
    lastModififed: str = ""

@dataclass
class TargetState:
    metadata: ObjectMeta = None
    spec: TargetSpec = None
    status: TargetStatus = None

@dataclass
class DeviceSpec:
    properties: Dict[str, str] = None
    bindings: List[BindingSpec] = None
    displayName: str = ""
    
@dataclass
class DeploymentSpec:
    solutionName: str = ""
    solution: SolutionState = None
    instance: InstanceSpec = None
    targets: Dict[str, TargetState] = None
    devices: List[DeviceSpec] = None
    assignments: Dict[str, str] = None
    componentStartIndex: int = -1
    componentEndIndex: int= -1
    activeTarget: str = ""

    def get_components_slice(self) -> []:
        if self.solution != None:
            if self.componentStartIndex >= 0 and self.componentEndIndex >= 0 and self.componentEndIndex > self.componentStartIndex:
                return self.solution.spec.components[self.componentStartIndex: self.componentEndIndex]
            return self.solution.spec.components
        return []

@dataclass
class ComparisionPack:
    desired: List[ComponentSpec]
    current: List[ComponentSpec]

class ProxyHost(object):
    app = None 
    mqtt_client = None
    def __init__(self,  get, apply, mqtt_broker="localhost", mqtt_port=1883, request_topic="coa-request", response_topic="coa-response"):
        self.app = Flask(__name__)
        self.apply = apply
        self.get = get
        self.app.add_url_rule('/instances', 'instances', self.__instances, methods=['GET', 'POST', 'DELETE'])        
        self.response_topic = response_topic
        # Initialize MQTT client
        self.mqtt_client = mqtt.Client()
        self.mqtt_client.on_connect = self.on_connect
        self.mqtt_client.on_message = self.on_message
        self.mqtt_client.connect(mqtt_broker, mqtt_port)
        
        # Subscribe to relevant topics
        self.mqtt_client.subscribe(request_topic)
        self.mqtt_client.loop_start()

    def run (self):
        serve(self.app, host='0.0.0.0', port=8090, threads=1)   

    def on_connect(self, client, userdata, flags, rc):
        print("Connected to MQTT Broker with result code " + str(rc))

    def on_message(self, client, userdata, msg):
        data = json.loads(msg.payload.decode())
        if msg.topic == "apply":
            response = self.__apply(data)
        elif msg.topic == "get":
            response = self.__get(data)
        else:
            response = {"error": "Unknown topic"}
        
        # Publish response to a specific response topic
        self.mqtt_client.publish(self.response_topic, json.dumps(response)) 
    def __instances(self):
        if request.method == 'POST':
            return self.__apply(request.get_json())
        elif request.method == 'DELETE':
            return self.__remove(request.get_json())
        elif request.method == 'GET':
            return self.__get(request.get_json())
        else:
            abort(405)
    def __apply(self, data):
        deployment = jsons.loads(json.dumps(data), DeploymentSpec)
        components = deployment.get_components_slice()    
        return self.apply(components)
    def __get(self,data):
        deployment = jsons.loads(json.dumps(data), DeploymentSpec)
        components = deployment.get_components_slice()  
        return self.get(components)     

