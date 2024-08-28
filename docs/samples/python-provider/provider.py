from sdk_poc import *
import jsons

def deploy(components):    
    for component in components:
        # deploy the component            
        print("deploying ", component.name)         
    return ""

def remove(components):
    for component in components:
        # deploy the component            
        print("removing ", component.name)         
    return ""

def get(components):    
    ret = []
    # popluate a ComponentSpec array    
    return jsons.dump(ret)

def needs_update(pack):    
    for desired in pack.desired:
        found = False
        for current in pack.current:
            if current.name == desired.name or current.name.startswith(desired.name):
                found = True
                break
        if not found:
            return ""
    abort(409)

def needs_remove(pack):
    for desired in pack.desired:
        for current in pack.current:
            if current.name == desired.name or current.name.startswith(desired.name):
                return ""
    abort(409)

host = ProxyHost(deploy, remove, get, needs_update, needs_remove)
host.run()