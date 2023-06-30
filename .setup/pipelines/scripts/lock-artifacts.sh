#!/bin/env bash
set -e
set -x

docker_images=(
  "symphony-api"
  "symphony-k8s"
)
helm_chart="helm/symphony"
chart_directory="symphony-extension/helm/symphony"

registry="symphonycr"
build_nmuber="$(Build.BuildNumber)"
chart_version="$(cat ${chart_directory}/Chart.yaml | grep "^version" | awk '{print $2}')"

## This locks all generated artifacts in ACR
for image in "${docker_images[@]}"; do
  echo "Locking docker image ${image}:${build_nmuber} in ${registry}"
  az acr repository update \
    --name "${registry}" \
    --image "${image}:${build_nmuber}" \
    --write-enabled false
done

echo "Locking helm chart ${helm_chart}:${chart_version} in ${registry}"
az acr repository update \
    --name "${registry}" \
    --image "${helm_chart}:${chart_version}" \
    --write-enabled false