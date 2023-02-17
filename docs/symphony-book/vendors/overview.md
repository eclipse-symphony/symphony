# Vendors

## Vendor configuration
Vendors are configured as part of the [host configuration file](../hosts/overview.md#host-configuration), under the ```vendors``` array under the top-level ```api``` element. The following example shows an example of a simpliest ```vendors.echo``` vendor, which returns a string when invoked:
```json
{
    "api": {
        "vendors" :[
            {
                "type": "vendors.echo",
                "route": "greetings",
                "managers": []
            }
        ]
    }
}
```
The above configuration snippet defines a ```vendors.echo``` vendor to be loaded and configured on the ```greetings``` route. Once this vendor is loaded, you can access it via ```http(s)://<server-address>:<server-port>/v1alpha2/greetings```.

A more complex vendor usually loads a number of [Managers](../managers/overview.md), each in turn loads one or more [Providers](../providers/overview.md), for example, the following configuraiton snippet defines a ```vendors.targets``` vendor, which loads a ```managers.symphony.targets``` manager, which loads a ```providers.state.k8s``` provider:
```json
{
    "type": "vendors.targets",
    "route": "targets",
    "managers": [
        {
            "name": "targets-manager",
            "type": "managers.symphony.targets",
            "properties": {
                "providers.state": "k8s-state"
            },
            "providers": {
                "k8s-state": {
                    "type": "providers.state.k8s",
                    "config": {
                        "inCluster": true
                    }
                }
            }
        }
    ]
```
