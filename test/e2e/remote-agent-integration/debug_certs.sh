#!/bin/bash

echo "=== Checking CA Secret in cert-manager namespace ==="
kubectl get secret client-cert-secret -n cert-manager -o yaml

echo -e "\n=== Checking CA Certificate Subject ==="
kubectl get secret client-cert-secret -n cert-manager -o jsonpath='{.data.ca\.crt}' | base64 -d | openssl x509 -text -noout | grep "Subject:"

echo -e "\n=== Checking Bundle ConfigMap ==="
kubectl get configmap symphony-clientca-key -n default -o yaml

echo -e "\n=== Checking Symphony API Pod logs for authentication errors ==="
kubectl logs -n default deployment/symphony-api --tail=50 | grep -i -E "(403|auth|cert|tls)"

echo -e "\n=== Checking Bundle Status ==="
kubectl describe bundle symphonyclientca-bundle

echo -e "\n=== Checking latest test certificates ==="
LATEST_DIR=$(ls -t /tmp/symphony-e2e-test-* 2>/dev/null | head -1)
if [ -n "$LATEST_DIR" ] && [ -d "$LATEST_DIR" ]; then
    echo "Found test directory: $LATEST_DIR"
    echo "=== CA Certificate Subject ==="
    openssl x509 -in "$LATEST_DIR/ca.pem" -text -noout | grep "Subject:"
    echo "=== Client Certificate Subject ==="
    openssl x509 -in "$LATEST_DIR/client.pem" -text -noout | grep "Subject:"
    echo "=== Client Certificate Issuer ==="
    openssl x509 -in "$LATEST_DIR/client.pem" -text -noout | grep "Issuer:"
    echo "=== Verify Certificate Chain ==="
    openssl verify -CAfile "$LATEST_DIR/ca.pem" "$LATEST_DIR/client.pem"
else
    echo "No test directories found"
fi
