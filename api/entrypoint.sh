#!/bin/bash

set -e # exit on error

# Start the web server in background
export USE_SERVICE_ACCOUNT_TOKENS=false
export SYMPHONY_API_URL=http://localhost:8082/v1alpha2/
./symphony-api -c ./symphony-api-no-k8s.json -l Debug &
SYMPHONY_PID=$!

until curl -s http://localhost:8082/v1alpha2/greetings > /dev/null; do
    echo "Waiting for symphony-api..."
    sleep 2
done

echo "symphony-api is up. Waiting for it to exit..."

# Wait for the symphony-api process to finish
wait $SYMPHONY_PID

echo "symphony-api exited."