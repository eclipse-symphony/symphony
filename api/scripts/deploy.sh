#!/bin/bash

set -e # exit on error


# Start the web server in background
export USE_SERVICE_ACCOUNT_TOKENS=false
export SYMPHONY_API_URL=http://localhost:8082/v1alpha2

TOKEN=$(curl -X POST -H "Content-Type: application/json" -d '{"username":"admin","password":""}' "${SYMPHONY_API_URL}/users/auth" | jq -r '.accessToken')

echo "Calling /greetings endpoint..."
GREETING_RESPONSE=$(curl -s -X GET -H "Authorization: Bearer $TOKEN" "${SYMPHONY_API_URL}/greetings")
echo "Greeting response: $GREETING_RESPONSE"

echo "Starting deployment..."
DEPLOY_RESPONSE=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" --data @/target.json "${SYMPHONY_API_URL}/targets/registry/pc-target")
echo "Deployment response: $DEPLOY_RESPONSE"
