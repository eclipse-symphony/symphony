# Troubleshooting
When running Symphony in Kubernetes, you can leverage familiar Kubernetes utilities to diagnose Symphony components and objects. This section provides some guidance on debugging common problems you may encounter.

## Check Symphony Health and Logs
When hosted on Kubernetes, Symphony runs several pods, such as the API pod (```symphony-api-*```), the controller pod (```symphony-controller-manager-*```), the Redis pod (```symphony-redis-*```) and the Zipkin pod (```symphony-zipkin-*```). The first thing to check is to see if these pods are healthy using ```kubectl get pods```.

If any of the pods has crashed or stopped, examine the log with:

```bash
Kubectl logs <pod name>
```
The log should give you a pretty good idea of where the API might have failed. If you found a pod got restarted for some reason, you can use the ```-f``` switch to observe the logs, try to make the pod crash again and observe the error message.

## Check Symphony object status
Objects like ```Solutions```, ```Catalogs```, ```Campaigns``` rarely have problems with themselves, as they don’t trigger additional system operations. On the other hand, ```Instances```, ```Targets```, ```Activations``` are often where errors are observed. When you see an object is at error state, you can dump the YAML document and observe recorded status, which should contain the error information. For example:
```bash
# examine an instance
kubectl get instance <instance name> -o yaml
# examine a target
kubectl get target <target name> -o yaml
```
> **NOTE**: Symphony runs continuous state reconciliation loops on ```Instances``` and ```Targets```. Sometimes an error may resolve itself over time. The default interval of reconciliation is about 3 minutes. 

## More Topics

* [Debugging Symphony API](./debugging-api.md)
