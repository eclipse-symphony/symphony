# Counter Campaign Sample

## Invoke via kubectl
```bash
kubectl apply -f campaign.yaml
kubectl apply -f activation.yaml
```
## Invoke via REST API
```bash
TOKEN=$(curl -X POST -H "Content-Type: application/json" -d '{"username":"admin","password":""}' http://localhost:8080/v1alpha2/users/auth | jq -r '.accessToken')
curl -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d @'./campaign-container.json' http://localhost:8080/v1alpha2/campaigncontainers/counter-campaign
curl -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d @'./campaign.json' http://localhost:8080/v1alpha2/campaigns/counter-campaign-v-version1
curl -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d @'./activation.json' http://localhost:8080/v1alpha2/activations/registry/counter-campaign-activation
```