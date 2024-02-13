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
class DeviceSpec:
    properties: Dict[str, str] = None
    bindings: List[BindingSpec] = None
    displayName: str = ""
    

@dataclass
class DeploymentSpec:
    solutionName: str = ""
    solution: SolutionSpec = None
    instance: InstanceSpec = None
    targets: Dict[str, TargetSpec] = None
    devices: List[DeviceSpec] = None
    assignments: Dict[str, str] = None
    componentStartIndex: int = -1
    componentEndIndex: int= -1
    activeTarget: str = ""

    def get_components_slice(self) -> []:
        if self.solution != None:
            if self.componentStartIndex >= 0 and self.componentEndIndex >= 0 and self.componentEndIndex > self.componentStartIndex:
                return self.solution.components[self.componentStartIndex: self.componentEndIndex]
            return self.solution.components
        return []

@dataclass
class ComparisionPack:
    desired: List[ComponentSpec]
    current: List[ComponentSpec]

class ProxyHost(object):
    app = None 
    def __init__(self,  apply, remove, get, needs_update, needs_remove):
        self.app = Flask(__name__)
        self.apply = apply
        self.remove = remove
        self.get = get
        self.needs_update = needs_update
        self.needs_remove = needs_remove     
        self.app.add_url_rule('/instances', 'instances', self.__instances, methods=['GET', 'POST', 'DELETE'])        
        self.app.add_url_rule('/needsupdate', 'needsupdate', self.__needs_update)        
        self.app.add_url_rule('/needsremove', 'needsremove', self.__needs_remove)        
    def run (self):
        serve(self.app, host='0.0.0.0', port=8090, threads=1)        
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
    def __remove(self, data):
        deployment = jsons.loads(json.dumps(data), DeploymentSpec)
        components = deployment.get_components_slice()
        return self.remove(components)
    def __get(self,data):
        deployment = jsons.loads(json.dumps(data), DeploymentSpec)
        components = deployment.get_components_slice()  
        return self.get(components)
    def __needs_update(self):
        pack = jsons.loads(json.dumps(request.get_json()), ComparisionPack)         
        return self.needs_update(pack)
    def __needs_remove(self):
        pack = jsons.loads(json.dumps(request.get_json()), ComparisionPack)         
        return self.needs_remove(pack)    
        

