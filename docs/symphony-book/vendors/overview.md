# Vendors

A vendor defines a chunk of API. You can consider each vendor a nano service, which can be assembled into one or multiple microservices. Developing with vendors uses an API-first approach, in which API shapes are mocked up and tested before any API implementations. This allows rapid iterations of APIs to satisfy project requirements before any expensive API implementation takes place.

## Vendor configuration

Vendors are configured as part of the [host configuration file](../hosts/overview.md#host-configuration), under the `vendors` array under the top-level `api` element. The following example shows an example of a simple `vendors.echo` vendor, which returns a string when invoked:

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

The previous configuration snippet defines a `vendors.echo` vendor to be loaded and configured on the `greetings` route. Once this vendor is loaded, you can access it via `http(s)://<server-address>:<server-port>/v1alpha2/greetings`.

A more complex vendor usually loads a number of [managers](../managers/overview.md), each in turn loads one or more [providers](../providers/_overview.md). For example, the following configuration snippet defines a `vendors.targets` vendor, which loads a `managers.symphony.targets` manager, which loads a `providers.state.k8s` provider:

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
}
```

## Publish and subscribe

Symphony doesn't allow any horizontal dependencies across vendors, managers or providers. Instead, these components can exchange messages with each other through a pub/sub system provided by Symphony vendors.

The vendor object has a `VendorContext` property. It has a `Publish` method and a `Subscribe` method for messaging. When a vendor needs to publish an event, it simply uses its context property to publish to a topic:

```go
c.Vendor.Context.Publish("trace", v1alpha2.Event{
    Body: "test message",
})
```

Similarly, a vendor can subscribe to a topic, for example:

```go
e.Vendor.Context.Subscribe("trace", func(topic string, event v1alpha2.Event) error {
    msg := event.Body.(string)
    fmt.Println(msg)
    return nil
})
```

When a vendor creates its Managers, it injects its context as the manager context. This allows Managers to publish/subscribe to events through the manager context as well.

## Pub/Sub at scale

By default, Symphony is configured to use an in-memory message bus. You can configure the in-memory message bus at either the vendor level or the host level. If the in-memory bus is configured at the vendor level, all events are scoped to the specific vendor, which means managers under the same vendor can communicate with each other, but not across vendors. If the in-memory bus is configured at the host level, all vendors hosted on the same Symphony process can message each other.

The in-memory message bus has two major shortcomings: first, it doesn't support cross-process messaging. Second, it doesn't provide guaranteed delivery. In a production environment, you probably want to configure a scalable messaging backend, such as Redis, instead of using an in-memory message bus.

> **NOTE**: Symphony is likely to have Redis configured as the default message bus before release.
