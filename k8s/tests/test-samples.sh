##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##
function check_instance {
    while : 
    do 
        deployed=$(kubectl get instance $1 -o=jsonpath='{.status.properties.deployed}')
        status=$(kubectl get instance $1 -o=jsonpath='{.status.properties.status}')
        if [ "$status" = "OK" ] && [ "$deployed" = "$2" ]
        then 
            echo "PASS"
            break
        fi
        if [ "$deployed" != "" ]
        then 
            echo "FAILED"
            break
        fi
        sleep 3
    done
}
function adaptive_deployment {
    echo "****************************************"
    echo "*       Testing Adaptive Deployment    *"
    echo "****************************************"
    
    kubectl create -f ../../symphony-docs/samples/scenarios/adaptive/iot-target.yaml &> /dev/null
    kubectl create -f ../../symphony-docs/samples/scenarios/adaptive/k8s-target.yaml &> /dev/null
    kubectl create -f ../../symphony-docs/samples/scenarios/adaptive/solution.yaml &> /dev/null
    kubectl create -f ../../symphony-docs/samples/scenarios/adaptive/instance.yaml &> /dev/null

    check_instance "my-instance" 2

    kubectl delete instance my-instance &> /dev/null    
    kubectl delete target basic-k8s-target &> /dev/null
    kubectl delete target voe-target &> /dev/null
    kubectl delete solution redis-server &> /dev/null
} 

function k8s_hello_world {
    echo "****************************************"
    echo "*       Testing K8s Hello World        *"
    echo "****************************************"

    kubectl create -f ../../symphony-docs/samples/k8s/hello-world/target.yaml &> /dev/null
    kubectl create -f ../../symphony-docs/samples/k8s/hello-world/solution.yaml &> /dev/null
    kubectl create -f ../../symphony-docs/samples/k8s/hello-world/instance.yaml &> /dev/null

    check_instance "redis-instance" 1

    kubectl delete instance redis-instance &> /dev/null
    kubectl delete solution redis-server &> /dev/null
    kubectl delete target basic-k8s-target &> /dev/null
}

function iot_simulated_sensor {
    echo "****************************************"
    echo "*  Testing IoT Edge Simulated Sensor   *"
    echo "****************************************"

    kubectl create -f ../../symphony-docs/samples/iot-edge/simulated-temperature-sensor/target.yaml &> /dev/null
    kubectl create -f ../../symphony-docs/samples/iot-edge/simulated-temperature-sensor/solution.yaml &> /dev/null
    kubectl create -f ../../symphony-docs/samples/iot-edge/simulated-temperature-sensor/instance-1.yaml &> /dev/null
    kubectl create -f ../../symphony-docs/samples/iot-edge/simulated-temperature-sensor/instance-2.yaml &> /dev/null

    check_instance "my-instance-1" 1
    check_instance "my-instance-2" 1

    kubectl delete instance my-instance-1 &> /dev/null
    kubectl delete instance my-instance-2 &> /dev/null    
    kubectl delete target voe-target-1 &> /dev/null
    kubectl delete solution simulated-temperature-sensor &> /dev/null
}

function k8s_symphony_agent {
    echo "****************************************"
    echo "*     Testing K8s Symphony Agent       *"
    echo "****************************************"

    kubectl create -f ../../symphony-docs/samples/k8s/symphony-agent/target.yaml &> /dev/null

    deployment=$(kubectl get deployment | grep target-runtime)
    svc=$(kubectl get svc | grep symphony-agent)

    if [ "$deployment" = "" ] || [ "$svc" = "" ]
    then
        echo "FAILED"
        return 
    fi

    kubectl delete target symphony-k8s-target &> /dev/null

    deployment=$(kubectl get deployment | grep target-runtime)
    svc=$(kubectl get svc | grep symphony-agent)

    if [ "$deployment" != "" ] || [ "$svc" != "" ]
    then
        echo "FAILED"
        return 
    fi

    echo "PASS"
}

k8s_hello_world
k8s_symphony_agent
adaptive_deployment
iot_simulated_sensor