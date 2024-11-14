#
# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
# SPDX-License-Identifier: MIT
#
from dataclasses import dataclass
from typing import List
from typing import Dict, Any
from flask import Flask, abort, request, Response
import jsons
import json
from waitress import serve
import paho.mqtt.client as mqtt

@dataclass
class ObjectMeta:
    namespace: str = ""
    name:  str = "" 
    eTag: str = ""
    objGeneration: int = 0
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
class InstanceSpec:
    displayName: str = ""
    scope: str = ""    
    parameters: Dict[str, str] = None
    metadata: Dict[str, str] = None
    solution: str
    target: TargetSelector = None
    topologies: List[TopologySpec] = None
    pipelines: List[PipelineSpec] = None
    isDryRun: bool = False

@dataclass
class FilterSpec:
    direction: str = ""
    type: str = ""
    parameters: Dict[str, str] = None
    
@dataclass
class RouteSpec:
    route: str = ""
    type: str = ""
    properties: Dict[str, str] = None
    filters: List[FilterSpec] = None

@dataclass
class SidecarSpec:
    name: str = ""
    type: str = ""
    properties: Dict[str, str] = None

@dataclass
class ComponentSpec:
    name: str = ""   
    type: str = ""
    metadata: Dict[str, str] = None
    properties: Dict[str, Any] = None
    parameters: Dict[str, str] = None
    routes: List[RouteSpec] = None
    constraints: str = ""
    depedencies: List[str] = None
    skills: List[str] = None
    sidecars: List[SidecarSpec] = None

@dataclass
class SolutionSpec:    
    displayName: str = ""
    metadata: Dict[str,str] = None
    components: List[ComponentSpec] = None    
    version: str = ""
    rootResource: str = ""
    
@dataclass
class SolutionState:
    metadata: ObjectMeta = None
    spec: SolutionSpec = None    

@dataclass
class TargetSpec:
    displayName: str = ""
    scope: str = ""
    metadata: Dict[str, str] = None
    properties: Dict[str, str] = None
    components: List[ComponentSpec] = None
    constraints: str = ""
    topologies: List[TopologySpec] = None
    forceRedeploy: bool = False
    isDryRun: bool = False

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
    percentComplete: int = 0
    failureCause: str = ""
    logErrors: bool = False
    error: ErrorType = None
    output: Dict[str, str] = None

@dataclass
class DeployableStatus:
    properties: Dict[str, str] = None
    provisioningStatus: ProvisioningStatus = None
    lastModififed: str = "" #TODO: use datetime.datetime instead?

@dataclass
class TargetState:
    metadata: ObjectMeta = None
    spec: TargetSpec = None
    status: DeployableStatus = None

@dataclass
class InstanceState:
    metadata: ObjectMeta = None
    spec: InstanceSpec = None
    status: DeployableStatus = None

@dataclass
class DeviceSpec:
    displayName: str = ""
    properties: Dict[str, str] = None
    bindings: List[BindingSpec] = None
    
@dataclass
class DeploymentSpec:
    solutionName: str = ""
    solution: SolutionState = None
    instance: InstanceState = None
    targets: Dict[str, TargetState] = None
    devices: List[DeviceSpec] = None
    assignments: Dict[str, str] = None
    componentStartIndex: int = -1
    componentEndIndex: int= -1
    activeTarget: str = ""
    generation: str = ""
    jobID: str = ""
    objectNamespace: str = ""
    hash: str = ""
    isDryRun: bool = False

    def get_components_slice(self) -> []:
        if self.solution != None:
            if self.componentStartIndex >= 0 and self.componentEndIndex >= 0 and self.componentEndIndex > self.componentStartIndex:
                return self.solution.spec.components[self.componentStartIndex: self.componentEndIndex]
            return self.solution.spec.components
        return []

@dataclass
class COARequest:
    method: str
    route: str
    contentType: str
    body: bytes = b"" 
    metadata: Dict[str, str] = None
    parameters: Dict[str, str] = None

@dataclass
class COAResponse:
    contentType: str 
    body: bytes = b"" 
    state: int
    metadata: Dict[str, str] = None
    redirectUri: str = ""

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

    def create_bad_request_response(message: str, metadata: Dict[str, str] = None) -> COAResponse:
        return COAResponse(
            contentType="application/json",
            body=json.dumps({"error": message}).encode(),
            state=400,
            metadata=metadata or {},
            redirectUri=""
        )
    def on_message(self, client, userdata, msg):
        data = json.loads(msg.payload.decode())
        coa_request = jsons.loads(json.dumps(data), COARequest)

        if coa_request.route == "solution/instances":
            namespace = request.parameters.get("namespace", "default")
            target_name = request.metadata.get("active-target", "") if request.metadata else ""
            deployment = json.loads(request.body.decode())

            if coa_request.method == "POST":            
                response = self.__apply(deployment)
            elif coa_request.method == "GET":
                response = self.__get(data)
            else:
                response = create_bad_request_response(f"Unknown method: {coa_request.method}")
        else:
            response = create_bad_request_response(f"Unknown route: {coa_request.route}")
        
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

