#!/bin/env bash
set -e
TIMEOUTDEPLOYMENT="15m"
TIMEOUT="1m"
TIMEOUTFINALIZER=60
SOLUTION_GROUP=solution.symphony
FABRIC_GROUP=fabric.symphony
AI_GROUP=ai.symphony
WORKFLOW_GROUP=workflow.symphony
FEDERATION_GROUP=federation.symphony
# Ensure the DEPLOYMENT_FINALIZER environment variable is set
if [ -z "$DEPLOYMENT_FINALIZER" ]; then
  echo "Error: DEPLOYMENT_FINALIZER environment variable is not set."
  exit 1
fi

function delete_crds_instances {
  echo "Patching finalizers for instances.$SOLUTION_GROUP and targets.$FABRIC_GROUP resources"
  patch_finalizers "instances.$SOLUTION_GROUP" $DEPLOYMENT_FINALIZER || true
  patch_finalizers "targets.$FABRIC_GROUP" $DEPLOYMENT_FINALIZER || true

  echo "Deleting instances.$SOLUTION_GROUP and targets.$FABRIC_GROUP resources"
  delete_crs "instances.$SOLUTION_GROUP"
  delete_crs "targets.$FABRIC_GROUP"
  
  kubectl delete crds "instances.$SOLUTION_GROUP" --wait --timeout=$TIMEOUT --ignore-not-found || true 
  if [ $? -ne 0 ]; then
    echo "Failed to delete CRD instances.$SOLUTION_GROUP, invoking remove_finalizers"
    remove_finalizers "instances.$SOLUTION_GROUP"
  fi
  echo "Deleting solutions.$SOLUTION_GROUP"
  kubectl delete crds "solutions.$SOLUTION_GROUP" --wait --timeout=$TIMEOUT --ignore-not-found || true 
  if [ $? -ne 0 ]; then
    echo "Failed to delete CRD solutions.$SOLUTION_GROUP, invoking remove_finalizers"
    remove_finalizers "solutions.$SOLUTION_GROUP"
  fi
  echo "Deleting targets.$FABRIC_GROUP"
  kubectl delete crds "targets.$FABRIC_GROUP" --wait --timeout=$TIMEOUT --ignore-not-found || true 
  if [ $? -ne 0 ]; then
    echo "Failed to delete CRD targets.$FABRIC_GROUP, invoking remove_finalizers"
    remove_finalizers "targets.$FABRIC_GROUP"
  fi
}

function delete_crds_campaigns {
  echo "Deleting activations.$WORKFLOW_GROUP"
  kubectl delete crds "activations.$WORKFLOW_GROUP" --wait --timeout=$TIMEOUT --ignore-not-found || true 
  if [ $? -ne 0 ]; then
    echo "Failed to delete CRD activations.$WORKFLOW_GROUP, invoking remove_finalizers"
    remove_finalizers "activations.$WORKFLOW_GROUP"
  fi

  echo "Deleting campaigns.$WORKFLOW_GROUP"
  kubectl delete crds "campaigns.$WORKFLOW_GROUP" --wait --timeout=$TIMEOUT --ignore-not-found || true 
  if [ $? -ne 0 ]; then
    echo "Failed to delete CRD campaigns.$WORKFLOW_GROUP, invoking remove_finalizers"
    remove_finalizers "campaigns.$WORKFLOW_GROUP"
  fi
}

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

# Function to patch the finalizers of a resource type to a specific value
patch_finalizers() {
  local resource_type=$1
  local finalizer_value=$2

  # Fetch all resources of the given type and patch the finalizers
  kubectl get "$resource_type" --all-namespaces -o jsonpath="{range .items[*]}{.metadata.namespace}{'\t'}{.metadata.name}{'\t'}{.metadata.finalizers}{'\n'}{end}" |
    while read -r namespace name finalizers; do
      if echo "$finalizers" | grep -q "$finalizer_value"; then
        echo "Finalizer $finalizer_value already present for $resource_type $name in namespace $namespace"
      else
        # Append the new finalizer to the existing finalizers
        new_finalizers=$(echo "$finalizers" | jq -c --arg finalizer "$finalizer_value" '. + [$finalizer]')
        echo "Patching finalizers for $resource_type $name in namespace $namespace"
        kubectl patch "$resource_type" "$name" -n "$namespace" --type='merge' -p "{\"metadata\":{\"finalizers\":$new_finalizers}}"
        if [ $? -ne 0 ]; then
          echo "Failed to patch finalizers for $resource_type $name in namespace $namespace"
        fi
      fi
    done
}

# Function to delete all resources of a given type
delete_crs() {
  local resource_type=$1

  # Fetch all resources of the given type and delete them
  kubectl get "$resource_type" --all-namespaces -o jsonpath="{range .items[*]}{.metadata.namespace}{'\t'}{.metadata.name}{'\n'}{end}" |
    while read -r namespace name; do
      echo "Deleting $resource_type $name in namespace $namespace"
      kubectl delete "$resource_type" "$name" -n "$namespace" --wait --timeout=$TIMEOUTDEPLOYMENT
    done
}

function remove_finalizers {
  # Get the resource type from the function argument
  local resource_type=$1

  # fetch all resources of the given type and patch the finalizers to an empty array
  patchResource "$resource_type" '{"metadata":{"finalizers":[]}}' &
  local patch_pid=$!

  sleep $TIMEOUTFINALIZER &
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

echo "Deleting Symphony resources"
# Use the function for each resource types in order

resource_types=(
  "devices.$FABRIC_GROUP"
  "models.$AI_GROUP"
  "skills.$AI_GROUP"
  "skillpackages.$AI_GROUP"
  "sites.$FEDERATION_GROUP"
  "catalogs.$FEDERATION_GROUP"
)

delete_crds_campaigns 

delete_crds_instances

for resource_type in "${resource_types[@]}"; do
    echo "Deleting $resource_type" 
    kubectl delete crds "$resource_type" --wait --timeout=$TIMEOUT --ignore-not-found || true 
    if [ $? -ne 0 ]; then
      echo "Failed to delete CRD $resource_type, invoking remove_finalizers"
      remove_finalizers "$resource_type" 
    fi
done

echo "All delete operations completed"