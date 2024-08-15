# State Providers Interface

## Upsert
Upsert a object to state store. Use UpsertOption to control the behavior of upsert.

## Get
Get an object from state store.

## Delete
Delete an object from state store.

## List
List objects from state store that meet the condition. Use FilterType and FilterValue to specify extra conditions.

### List Filters
When you query Symphony objects, you can attach an optional filter. 

#### Label filter
You can use a label filter to filter objects by object metadata labels. For example, to list only objects with a label `foo` equals `bar`, you can use the filter:
```json
{
    "filterType": "label",
    "filterValue": "foo==bar,foo2=bar2"
}
```
Label filter functions in the same way as [Kubernetes label selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/).
#### Field filter
You can use a field selector to filter objects based on some object metadata values and state values (some restrictions apply, see the link below). For example, to list only objects with a metadata `name` equals `c2`, you can use the filter:
```json
{
    "filterType": "field",
    "filterValue": "metadata.name=c2"
}
```
Field filter functions in the same way as [Kubernetes field selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/).
#### Spec filter 
You can use a JsonPath expression to filter objects by a spec field. For example, to list only specs with a `foo` property equals to `bar`, you can use the filter:
```json
{
    "filterType": "spec",
    "filterValue": "[?(@.properties.foo==\"bar\")]"
}
```
#### Status filter 
You can use a JsonPath expression to filter objects by a status field. For example, to list only specs with a status `properties.foo` field equals to `bar`, you can use the filter:
```json
{
    "filterType": "status",
    "filterValue": "[?(@.properties.foo==\"bar\")]"
}
```