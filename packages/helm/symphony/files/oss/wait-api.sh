#!/bin/sh
set -e

TIMEOUT=${TIMEOUT:-"300"} # default timeout is 300 seconds
# convert the timeout to an integer
TIMEOUT=$(echo $TIMEOUT | awk '{print int($1)}')

# poll the symphony-api deployment's container "symphony-api" container's image name equals to $SYMPHONY_API_IMAGE and "symphony-api" deployment in ready state
echo "Waiting for Symphony API deployment to be ready, image name: $SYMPHONY_API_IMAGE, namespace: $SYMPHONY_API_NAMESPACE, timeout is $TIMEOUT seconds."
end=$((SECONDS + TIMEOUT))

while [ $SECONDS -lt $end ]; do
    if kubectl get deployment symphony-api -n $SYMPHONY_API_NAMESPACE -o jsonpath='{.spec.template.spec.containers[?(@.name=="symphony-api")].image}' | grep -q "$SYMPHONY_API_IMAGE"; then
        echo "Fetching image name $SYMPHONY_API_IMAGE found."
        # check if the deployment is in a ready state
        if kubectl rollout status deployment/symphony-api -n $SYMPHONY_API_NAMESPACE -w --timeout=5s | grep -q "successfully rolled out"; then
            echo "Deployment symphony-api is ready."
            exit 0
        else
            echo "Deployment symphony-api is not ready yet, retrying..."
            sleep 5
        fi
    fi
  echo "symphony-api deployment with container image name $SYMPHONY_API_IMAGE not found, retrying..."
  sleep 5
done

echo "Cannot wait for Symphony API deployment to be ready, image name: $SYMPHONY_API_IMAGE. Still exit with 0 and not interrupt the normal deployment."
exit 0