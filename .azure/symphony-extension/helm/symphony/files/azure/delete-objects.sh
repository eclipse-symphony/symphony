#!/bin/env bash
set -e
TIMEOUT="60s"
GROUP=symphony.microsoft.com

function delete_crds {
  local resource_type=$1

  echo "Deleting $resource_type"
  kubectl delete crds "$resource_type" --wait --timeout=$TIMEOUT --ignore-not-found
}

echo "Deleting Symphony resources"
# Use the function for each resource types in order
delete_crds "instances.$GROUP"
delete_crds "solutions.$GROUP"
delete_crds "targets.$GROUP"
