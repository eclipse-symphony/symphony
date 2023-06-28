#!/bin/env bash
set -e
BUILD_NUMBER="$(Build.BuildNumber)"

pushd "k8s" || exit 1
$HOME/go/bin/mage helmTemplate
popd || exit 1

pushd "symphony-extension/helm/symphony" || exit 1

# Update values.yaml
sed -i "s/\(\s*tag: \).*/\1\"$BUILD_NUMBER\"/" values.yaml

# Update Chart.yaml
sed -i "s/\(^appVersion: \).*/\1\"$BUILD_NUMBER\"/" Chart.yaml
sed -i "s/\(^version: \).*/\1\"$BUILD_NUMBER\"/" Chart.yaml