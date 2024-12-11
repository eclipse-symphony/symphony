# Truck Fleet Management

## Prepare Demo Artifacts
### Deploy Fuel Pump App
```bash
# docs/samples/truck-fleet/fuel-pump
kubectl apply -f fuel-pump-deployment.yaml
# when service is ready
kubectl port-forward svc/fuel-pump-service 8000:8000
```
### Deploy Symphony Opera Portal
```bash
# docs/samples/truck-fleet/portal
kubectl apply -f opera-deployment.yaml
# when service is ready
kubectl port-forward svc/opear-service 3000:3000
```

### Deploy Symphony Objects
```bash
# docs/samples/truck-fleet/
kubectl apply -f targets/
kubectl apply -f workflows/docking.yaml
```

### Set up port-forwards
If you are using Minikube, you need to enable a few port forwards:
```bash
kubectl port-forward svc/opera-service 3000:3000 # Opera portal
kubectl port-forward svc/symphony-service  8080:8080 # Symphony service
kubectl port-forward svc/fuel-pump-service 8000:8000 # Fuel Pump service
```

## Demo Steps
1. Launch truck detector
    ```powershell
    # docs/samples/truck-fleet/truck-detection
    python .\truck_detector.py
    ```
2. Observe
    * New `avl-truck` Target is created on Opera portal
    * An email is sent notifying the driver is too tired
    * Truck parks at the pumping station

3. Click on the **Fuel up** button on the pumping station UI
4. Observe fuel is filled up on both pumping station UI and Opera portal (manual referesh needed)

## Clean up
1. Delete `avl-truck` Target:
    ```bash
    kubectl delete target avl-truck
    ```
2. Delete Activation:
    ```bash
    kubectl delete activation box-truck-docking-activation
    ```
3. Re-deploy Fuel Pump app to reset its state.
4. Stop the truck detector app.