#!/bin/env bash

# This script is used to remove finalizers from the Symphony Custom Resources
# This is needed incase the Symphony Operator is not running and the CRs are stuck in a terminating state

TIMEOUT=60
GROUP=symphony.microsoft.com

patchResource() {
  local resource_type="$1"
  local patch_data="$2"

  kubectl get "$resource_type" --all-namespaces -o jsonpath="{range .items[*]}{.metadata.namespace}{'\t'}{.metadata.name}{'\n'}{end}" |
    while read -r namespace name; do
      echo "Removing finalizers from $resource_type $name in namespace $namespace"
      kubectl patch "$resource_type" "$name" -n "$namespace" -p "$patch_data" --type=merge
      if [ $? -ne 0 ]; then
        echo "Failed to remove finalizers from $resource_type $name in namespace $namespace"
      fi
    done
}

function remove_finalizers {
  # Get the resource type from the function argument
  local resource_type=$1

  # fetch all resources of the given type and patch the finalizers to an empty array
  patchResource "$resource_type" '{"metadata":{"finalizers":[]}}' &
  local patch_pid=$!

  sleep $TIMEOUT &
  local sleep_pid=$!

  # Wait for the patch operation to finish or timeout to elapse
  while kill -0 $patch_pid 2>/dev/null && kill -0 $sleep_pid 2>/dev/null; do
    echo "Waiting ..."
    sleep 1
  done

  if kill -0 $patch_pid 2>/dev/null; then
    echo "$resource_type patch operation timed out"
    kill -9 $patch_pid
    return 1
  else
    echo "$resource_type patch operation completed"
    kill $sleep_pid 2>/dev/null
  fi
}
echo "Removing finalizers from Symphony resources"

remove_finalizers "instances.$GROUP"
remove_finalizers "targets.$GROUP"
