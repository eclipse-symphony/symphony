go build -o ./symphony-api
SYMPHONY_API_URL="http://localhost:8082/v1alpha2/" USE_SERVICE_ACCOUNT_TOKENS="false" ./symphony-api -c ./symphony-api-no-k8s.json -l Info
