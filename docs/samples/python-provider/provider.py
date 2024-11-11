from sdk_poc import *
import jsons

def apply(components):    
    for component in components:
        # deploy the component            
        print("deploying ", component.name)         
    return ""

def get(components):    
    ret = []
    # popluate a ComponentSpec array    
    return jsons.dump(ret)

host = ProxyHost(get, apply)
host.run()