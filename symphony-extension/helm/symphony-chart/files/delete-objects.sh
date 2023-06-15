#!/bin/env bash

TIMEOUT=60
GROUP=symphony.microsoft.com

function delete_resources {
  # Get the resource type from the function argument
  local resource_type=$1

  # Start the delete operation in the background and save the PID
  kubectl delete "$resource_type" --all -A --wait &
  local delete_pid=$!

  sleep $TIMEOUT &
  local sleep_pid=$!

  # Wait for the delete operation to finish or timeout to elapse
  while kill -0 $delete_pid 2>/dev/null && kill -0 $sleep_pid 2>/dev/null; do
    sleep 1
  done

  if kill -0 $delete_pid 2>/dev/null; then
    echo "$resource_type delete operation timed out"
    kill $delete_pid
    return 1
  else
    echo "$resource_type delete operation completed"
    kill $sleep_pid 2>/dev/null
  fi
}
echo "Deleting Symphony resources"
# Use the function for each resource types in order
delete_resources "instances.$GROUP"
delete_resources "solutions.$GROUP"
delete_resources "targets.$GROUP"
