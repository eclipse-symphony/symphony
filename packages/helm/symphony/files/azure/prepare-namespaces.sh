#!/bin/env bash
set -e

IFS="," read -ra LIST <<< "$NAMESPACES"

for namespace in "${LIST[@]}"; do
  if ! kubectl get namespace "$namespace"; then
    kubectl create namespace "$namespace"
  fi
done