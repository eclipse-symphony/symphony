# Catalog

A Symphony catalog is a piece of indexed information. It can contain a key-value pair collection itself, or it can be used as an index of a piece of external data, such as a document at a URL or a record in a database. Symphony saves Catalog objects in a non-SQL store and puts a graph engine on top to provide graph query capabilities.

The following is an example of a simple Symphony `catalog` that holds information about an office as a collection of key-value pairs. You can put any key-value pairs in a `catalog` object. This object represents a node in a graph.

```yaml
apiVersion: federation.symphony/v1
kind: Catalog
metadata:
  name: hq
spec:  
  type: asset
  properties:
    name: HQ
    address: 1 Microsoft Way
    city: Redmond
    state: WA
    zip: "98052"
    country: USA
    phone: "425-882-8080"
    version: "0.45.1"
    lat: "43.67961"
    lng: "-122.12826"
```

To represent a collection of edges, you can use an `edge` catalog:

> **NOTE**: In the current version, all edges are assumed to be directional, pointing from the key to the value in a key-value pair.

```yaml
apiVersion: federation.symphony/v1
kind: Catalog
metadata:
  name: edges
spec:  
  type: edge
  properties:
    node1: node2
    node2: node3
```

Symphony also provides an easier way to construct tree views with a `parentName` property, which can be used to point to a parent node/Catalog.

In any of the property values, you can refer to another catalog using a `<catalog-name>` expression. In the following example, the `line` property refers to another catalog object named `line-config`. All properties from the `line-config` object will be copied into the `line` property as child attributes. The sample also shows how you can use a `parentName` to set the parent node to a `global-config` catalog.

```yaml
apiVersion: federation.symphony/v1
kind: Catalog
metadata:
  name: app-config
spec:  
  type: config
  parentName: global-config
  metadata:
    asset: use-case
  properties:
    line: <line-config>    
```