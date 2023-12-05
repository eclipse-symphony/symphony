##
## Copyright (c) Microsoft Corporation.
## Licensed under the MIT license.
## SPDX-License-Identifier: MIT
##
mkdir -p $HOME/.symphony/k8s
mkdir -p $HOME/.symphony/iot-edge

cp -r ../docs/samples/k8s/hello-world/ $HOME/.symphony/k8s/hello-world/
cp -r ../docs/samples/k8s/staged/ $HOME/.symphony/k8s/staged/
cp -r ../docs/samples/iot-edge/simulated-temperature-sensor/ $HOME/.symphony/iot-edge/simulated-temperature-sensor/
cp ../docs/samples/samples.json $HOME/.symphony/
cp ../api/symphony-api $HOME/.symphony/
cp ../api/symphony-api-no-k8s.json $HOME/.symphony/