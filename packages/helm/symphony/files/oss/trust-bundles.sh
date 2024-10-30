#!/bin/sh
set -e

TIMEOUT=${TIMEOUT:-"300"} # default timeout is 300 seconds
# convert the timeout to an integer
TIMEOUT=$(echo $TIMEOUT | awk '{print int($1)}')

# label the namespace for trust bundle
echo "Labeling namespaces $TRUST_BUNDLE_NS with label $TRUST_BUNDLE_NS_LABEL_KEY=$TRUST_BUNDLE_NS_LABEL_VALUE"
kubectl label namespaces $TRUST_BUNDLE_NS $TRUST_BUNDLE_NS_LABEL_KEY=$TRUST_BUNDLE_NS_LABEL_VALUE

# wait for the trust bundle configmap to be populated
echo "Waiting for ConfigMap $TRUST_BUNDLE_CONFIGMAP_NAME to be populated, timeout is $TIMEOUT seconds."
end=$((SECONDS + TIMEOUT))
while [ $SECONDS -lt $end ]; do
  if kubectl get configmap $TRUST_BUNDLE_CONFIGMAP_NAME -n $TRUST_BUNDLE_NS > /dev/null 2>&1; then
    echo "ConfigMap $TRUST_BUNDLE_CONFIGMAP_NAME found."
    exit 0
  fi
  echo "ConfigMap $TRUST_BUNDLE_CONFIGMAP_NAME not found, retrying..."
  sleep 5
done

echo "Timed out waiting for ConfigMap $TRUST_BUNDLE_CONFIGMAP_NAME."
exit 1