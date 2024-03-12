##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##

# ----------------------------------------

# To run coa/pkg/apis/v1alpha2/bindings/http/jwt_test.go - TestAzureToken:

# Use the following command to genreate a new token, and then set AAD_TOKEN to the token value:
#   curl -X POST -H 'Content-Type: application/x-www-form-urlencoded' \
#   -d 'grant_type=client_credentials&client_id=<client id>&resource=2ff814a6-3304-4ab8-85cb-cd0e6f879c1d&client_secret=<client secret>' \
#   https://login.microsoftonline.com/<tenant id>/oauth2/token

# export AAD_TOKEN=<token value from above command>

# ----------------------------------------

# To run coa/pkg/apis/v1alpha2/bindings/http/jwt_test.go - TestRoleAssignment and TestRoleAssignmentRBACDisabled:

# Use Symphony API to log in as an administrator, and then set TEST_SYMPHONY_TOKEN to the token value:
# Note the user name has to be "admin"
#   curl -X POST -H 'Content-Type: application/json' -d '{"username": "admin", "password": ""}' http://localhost:8080/v1alpha2/users/auth

# export TEST_SYMPHONY_TOKEN=<accessToken field in return body>

# ----------------------------------------

# To run coa/pkg/apis/v1alpha2/cloudutils/azure/azureutils_test.go - TestGetTokenNoSecret:

# export TEST_AZURE_TENANT_ID=<Azure service principal tenant id>
# export TEST_AZURE_CLIENT_ID=<Azure service principal client id>

# ----------------------------------------

# To run coa/pkg/apis/v1alpha2/cloudutils/azure/azureutils_test.go - TestGetADUGroup:

# export TEST_AZURE_TENANT_ID=<Azure service principal tenant id>
# export TEST_AZURE_CLIENT_ID=<Azure service principal client id>
# export TEST_AZURE_CLIENT_SECRET=<Azure servcie principal client secret>
# export TEST_ADU_ACCOUNT_ENDPOINT=<ADU account endpoint>
# export TEST_ADU_ACCOUNT_INSTANCE=<ADU account instance>
# export TEST_ADU_GROUP=<ADU group>

# ----------------------------------------

# To run coa/pkg/apis/v1alpha2/cloudutils/azure/azureutils_test.go - TestGetADUDeployment and TestUpdateADUDeployment:

# export TEST_AZURE_TENANT_ID=<Azure service principal tenant id>
# export TEST_AZURE_CLIENT_ID=<Azure service principal client id>
# export TEST_AZURE_CLIENT_SECRET=<Azure servcie principal client secret>
# export TEST_ADU_ACCOUNT_ENDPOINT=<ADU account endpoint>
# export TEST_ADU_ACCOUNT_INSTANCE=<ADU account instance>
# export TEST_ADU_GROUP=<ADU group>
# export TEST_ADU_DEPLOYMENT=<ADU deployment>
# export TEST_ADU_UPDATE_PROVIDER=<ADU update provider used in the above deployment>

# ----------------------------------------

# To run coa/pkg/apis/v1alpha2/providers/probe/rtsp/rtsp_test.go - TestProbe:
# You'll need ffmpeg installed on your machine and then set TEST_RTSP to an open (not password protected) RTSP stream address, like "rtsp://20.212.158.240/1.mkv"
#   sudo apt-get install ffmpeg

# export TEST_RTSP=<RTSP stream address>

# ----------------------------------------

# To run coa/pkg/apis/v1alpha2/providers/pubsub/redis/redis_test.go - TestInit, TestBasicPubSub, TestBasicPubSubTwoProviders, TestMultipleSubscriber:
# You'll need a Redis server running at loalhost:6379 (by running a Redis container, for instance), and:

# export TEST_REDIS=yes

# ----------------------------------------

# To run coa/pkg/apis/v1alpha2/providers/reference/customvision/customvision_test.go - TestGet:
# You'll need Custom Vision configured and set the following:

# export TEST_CV_API_KEY=<CV API key>
# export TEST_CV_PROJECT=<CV project>
# export TEST_CV_ENDPOINT=<CV endpoint>
# export TEST_CV_ITERATION=<CV iteration>	

# ----------------------------------------

# To run coa/pkg/apis/v1alpha2/providers/reference/k8s/k8sprovider_test.go - TestInit:
# You'll need a Kubernetes cluster and your kubectl default context set to the cluster. Then, set:

# export TEST_K8S=yes

# ----------------------------------------

# To run coa/pkg/apis/v1alpha2/providers/reference/k8s/k8sprovider_test.go - TestGet:
# You'll need a Kubernetes cluster and your kubectl default context set to the cluster. 
# You'll also need Symphony deployed to the cluster and a Symphony solution created under the default namespace. Then, set:

# export TEST_K8S=yes
# export SYMPHONY_SOLUTION=<Symphony solution name>

# ----------------------------------------

# To run coa/pkg/apis/v1alpha2/providers/reporter/k8s/k8sreporter_test.go - TestInit:
# You'll need a Kubernetes cluster and your kubectl default context set to the cluster. Then, set:

# export TEST_K8S=yes

# ----------------------------------------

# To run coa/pkg/apis/v1alpha2/providers/reporter/k8s/k8sreporter_test.go - TestGet:
# You'll need a Kubernetes cluster and your kubectl default context set to the cluster. 
# You'll also need Symphony deployed to the cluster and a Symphony device created under the default namespace. Then, set:

# export TEST_K8S=yes
# export SYMPHONY_DEVICE=<Symphony device name>

# ----------------------------------------

# To run coa/pkg/apis/v1alpha2/providers/states/httpstate/httpstate_test.go - TestUpSert, TestList, TestDelete, TestGet:
# You'll need a Dapr sidecar running at http://localhost:3500, and set:

# export TEST_DAPR=yes

# ----------------------------------------

# To Test IoT Hub Target Provider, set these environment variables
# To create test environment:
#   1. Provision a new IoT Hub with default settings.
#   2. Add a new IoT Edge device to the hub. This is the "null" device referred as S8CTEST_IOTHUB_NULL_DEVICE_NAME.
#   3. Add another IoT Edge device with a simulated temperature module: https://docs.microsoft.com/en-us/azure/iot-edge/quickstart-linux?view=iotedge-2020-11. This is the "vm" device referred as S8CTEST_IOTHUB_VM_DEVICE_NAME
# export S8CTEST_IOTHUB_NAME=s8c-hub2.azure-devices.net
# export S8CTEST_IOTHUB_NULL_DEVICE_NAME=p8c-null
# export S8CTEST_IOTHUB_VM_DEVICE_NAME=p8c-vm
# export S8CTEST_IOTHUB_API_VERSION=2020-05-31-preview
# export S8CTEST_IOTEDGE_API_VERSION=2018-06-30
# export S8CTEST_IOTHUB_KEY_NAME=iothubowner
# export S8CTEST_IOTHUB_KEY=<IoT Hub Key>

# ------------------------------

# export P8CTEST_IOTHUB_NAME=p8c-hub.azure-devices.net
# export P8CTEST_IOTHUB_DEVICE_NAME=devkit-shadow
# export P8CTEST_IOTHUB_API_VERSION=2020-05-31-preview
# export P8CTEST_IOTEDGE_API_VERSION=2018-06-30
# export P8CTEST_IOTHUB_KEY_NAME=iothubowner
# export P8CTEST_IOTHUB_KEY=<IoT Hub Key>

# To test Custom Vision, set these environment variables
# export P8CTEST_CV_API_KEY=<CV API Key>
# export P8CTEST_CV_PROJECT=c781fb25-583e-4b8c-8cc4-49c9f6d4c47d
# export P8CTEST_CV_ENDPOINT=westus2.api.cognitive.microsoft.com
# export P8CTEST_CV_ITERATION=610f8e72-ead8-453e-ab65-07ac2edf6c08

# To test K8s, set the following environment varialbe, and make sure you have a my-devkit target
# export TEST_K8s=yes

# To test Win10 sideload, set TEST_WIN10_SIDELOAD to yes, and run test cases on Windows. Make sure you have app artifact in place

# To test Azure auth injector
# export CLIENT_ID=<client id>
# export CLIENT_SECRET=<client secret>
# export TENANT_ID=<tenant id>

total_count=0
total_succeeded=0
total_skipped=0

function test_folder {
    echo "**************************************************************"
    echo " Testing $1"
    echo "**************************************************************"
    echo
    exec 5>&1
    var=$(go test $1 -v -parallel 1 | tee /dev/fd/5)
    c1=$(echo "$var" | grep -c ' RUN') 
    c2=$(echo "$var" | grep -c ' PASS') 
    c3=$(echo "$var" | grep -c ' SKIP') 
    printf '\n# of test cases:    %d\n' "$c1"
    printf '# of passed cases:  %d\n' "$c2"
    printf '# of skipped cases: %d\n\n' "$c3"
    total_count=`expr $c1 + $total_count`
    total_succeeded=`expr $c2 + $total_succeeded`
    total_skipped=`expr $c3 + $total_skipped`
}

suite="${1:all}"



if [ "$suite" = "all" ] || [ "$suite" = "coa" ]
then 
    echo "*************************"
    echo "*    COA Unit Tests     *"
    echo "*************************"
    echo

    test_folder "../pkg/apis/v1alpha2/bindings/http"
    test_folder "../pkg/apis/v1alpha2/cloudutils/azure"
    test_folder "../pkg/apis/v1alpha2/observability"
    test_folder "../pkg/apis/v1alpha2/providers/certs/autogen"
    test_folder "../pkg/apis/v1alpha2/providers/certs/localfile"
    test_folder "../pkg/apis/v1alpha2/providers/probe/rtsp"    
    test_folder "../pkg/apis/v1alpha2/providers/pubsub/memory"
    test_folder "../pkg/apis/v1alpha2/providers/pubsub/redis"
    test_folder "../pkg/apis/v1alpha2/providers/reference/customvision"
    test_folder "../pkg/apis/v1alpha2/providers/reference/k8s"
    test_folder "../pkg/apis/v1alpha2/providers/reference/mock"
    test_folder "../pkg/apis/v1alpha2/providers/reporter/k8s"
    test_folder "../pkg/apis/v1alpha2/providers/states/httpstate"
    test_folder "../pkg/apis/v1alpha2/providers/states/memorystate"
    test_folder "../pkg/apis/v1alpha2/providers/uploader/azure/blob"
fi


if [ "$suite" = "all" ] || [ "$suite" = "pai" ]
then
    echo "**********************************"
    echo "*    Symphony API Unit Tests     *"
    echo "**********************************"
    echo
    pushd .
    cd ../../symphony-api
    go mod vendor
    test_folder "./pkg/apis/v1alpha1/managers/reference"
    test_folder "./pkg/apis/v1alpha1/managers/solution"
    test_folder "./pkg/apis/v1alpha1/providers/states/k8s"
    test_folder "./pkg/apis/v1alpha1/providers/target/azure/adu"
    test_folder "./pkg/apis/v1alpha1/providers/target/azure/iotedge"
    test_folder "./pkg/apis/v1alpha1/providers/target/helm"
    test_folder "./pkg/apis/v1alpha1/providers/target/http"
    test_folder "./pkg/apis/v1alpha1/providers/target/k8s"
    test_folder "./pkg/apis/v1alpha1/providers/target/kubectl"
    test_folder "./pkg/apis/v1alpha1/providers/target/proxy"
    test_folder "./pkg/apis/v1alpha1/providers/target/staging"
    test_folder "./pkg/apis/v1alpha1/providers/target/win10/sideload"
    test_folder "./pkg/apis/v1alpha1/utils"
    test_folder "./pkg/apis/v1alpha1/vendors"
    popd
fi

if [ "$suite" = "all" ] || [ "$suite" = "p8c" ]
then
    echo "*************************************"
    echo "*    Symphont K8s Binding Tests     *"
    echo "*************************************"
    echo
    pushd . 
    cd ../../symphony-k8s
    go mod vendor
    exec 5>&1
    var=$(make test | tee /dev/fd/5)
    c1=$(echo "$var" | grep -c ' RUN') 
    c2=$(echo "$var" | grep -c ' PASS') 
    c3=$(echo "$var" | grep -c ' SKIP') 
    printf '\n# of test cases:    %d\n' "$c1"
    printf '# of passed cases:  %d\n' "$c2"
    printf '# of skipped cases: %d\n\n' "$c3"
    total_count=`expr $c1 + $total_count`
    total_succeeded=`expr $c2 + $total_succeeded`
    total_skipped=`expr $c3 + $total_skipped`
    popd
fi

printf '\nTEST SUMMARY\n\n'
printf 'Total # of test cases:    %d\n' "$total_count"
printf 'Total # of passed cases:  %d\n' "$total_succeeded"
printf 'Total # of skipped cases: %d\n' "$total_skipped"
printf 'Total # of failed cases:  %d\n\n', `expr $total_count - $total_succeeded - $total_skipped`