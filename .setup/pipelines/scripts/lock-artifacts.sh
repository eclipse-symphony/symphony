#!/bin/env bash
set -e

ARTIFACTS=(
  "symphony-api"
  "symphony-k8s"
  "helm/symphony"
)
REGISTRY="symphonycr"
BUILD_NUMBER="$(Build.BuildNumber)"

## This locks all generated artifacts in ACR
for ARTIFACT in "${ARTIFACTS[@]}"; do
  az acr repository update \
    --name "${REGISTRY}" \
    --image "${ARTIFACT}:${BUILD_NUMBER}" \
    --write-enabled false
done
