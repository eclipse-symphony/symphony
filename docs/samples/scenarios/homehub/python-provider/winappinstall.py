from sdk_poc import *
import json
import subprocess
import re
import os

WINAPPINSTALL = "c:\\Program Files (x86)\\Windows Kits\\10\\bin\\10.0.22000.0\\x64\\WinAppDeployCmd.exe"
TARGET_IP = "192.168.50.29"

def deploy(components):    
    for component in components:
        if "app.package.path" in component.properties:
            cmd = WINAPPINSTALL + " install -ip " + TARGET_IP + " -file " + component.properties["app.package.path"]
            res = subprocess.call(cmd)
            if res != 0:
                abort(500)
    return ""

def get_packages(components, remove_postfix = True):

    cmd = WINAPPINSTALL + " list -ip " + TARGET_IP

    returned_output = subprocess.check_output(cmd)

    p = re.compile('^(\w+\.)+\w+$')
    ret = []
    
    for l in returned_output.splitlines():
        line = l.decode()
        if p.match(line) != None:
            m_line = line
            if remove_postfix:
                try:
                    u = m_line.rindex("__")
                except ValueError:
                    u = -1
                if u > 0:
                    m_line = line[:u]
                for c in components:
                    if c.name == m_line:
                        ret += [ComponentSpec(name=m_line, type="win.uwp")]
            else:
                for c in components:
                    if m_line.startswith(c.name):
                        ret += [ComponentSpec(name=m_line, type="win.uwp")]
    return ret

def remove(components):
    packages = get_packages(components, remove_postfix = False)
    for component in components:
        for ref in packages:
            if ref.name == component.name or ref.name.startswith(component.name):
                cmd = WINAPPINSTALL + " uninstall -ip " + TARGET_IP + " -package " + ref.name
                res = subprocess.call(cmd)
                if res != 0:
                    abort(500)
    return ""

def get(components):    
    return jsons.dump(get_packages(components))

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