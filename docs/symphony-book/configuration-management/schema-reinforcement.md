# Schema reinforcement

Symphony Catalog object uses an open key-value pair schema, which provides the foundation for Symphony to carry pieces of arbitrary information as a property bag. These information snippets can then be assembled into various graphs to represent larger information ontologies like asset trees, BOMs, sites and others.

In the case where a stronger schema check is required – just as limiting a configuration field to a certain value range – Symphony allows a Catalog to be annotated with a `schema` metadata that points to a schema definition. Once an Catalog is annotated with a schema, it will be checked against the schema on any update operations – regardless if you are using the REST API or using K8s API calls. Any schema violations will cause the update to be rejected.

## Schema rules

### Type check
Checks the type of field. For example, the following rule specifies `some-field` must be an integer:

```json
{
    "rules": {
        "some-field": {
            "type": "int"
        }
    }
}
```
Other value types include `float`, `bool`, `uint`, and `string`.

### Required field

Indicates if a field is mandatory:

```json
{
    "rules": {
        "some-field": {
            "required": true
        }
    }
}
```

### Regular expression

Specifies that a field must match a regular expression:

```json
{
    "rules": {
        "some-field": {
            "pattern": "^[a-z]+$"
        }
    }
}
```

Symphony also defines several pattern shortcuts for common regular expression patterns, as summarized in the following table:

|Shortcut| Expression|
|--------|--------|
|`<cidr>`| a valid CIDR range |
| `<dns-label>`| a valid DNS label|
| `<dns-name>` | a valid DNS name|
| `<email>` | a valid e-Mail address|
|`<ip4>`| a valid IP-v4 address|
|`<ip4-range>`|a valid IP-v4 address range|
|`<ip6>`|a valid IP-v6 address|
|`<ip6-range>`|a valid IP-v6 address range|
|`<mac-address>`| a valid MAC address|
|`<port>`|a valid port number|
| `<url>` | a valid HTTP(s) URL |
| `<uuid>` | an UUID |

### Symphony expression
You can use Symphony expressions to specify complex conditions, such as a field value has to fall into specific range:

```json
{
    "rules": {
        "some-field": {
            "pattern": "${{$and($gt($val(),10),$lt($val(),20))}}"
        }
    }
}
```
Some examples:

|Expression| Rule|
|--------|--------|
|`${{$and($gt($val(),10),$lt($val(),20))}}`| Value has to between 10 and 20 |
|`${{$in($val(), 'foo', 'bar', 'baz')}}`| Value has to be `foo`, `bar` or `baz` |

> **NOTE:** When you have multiple checks specified in a rule, they are applied at the same time. For example, you can specify a field being an integer, mandatory, and has to fall between certain range.


## Schema syntax for nested properties

Schema supports nested organizations using jq query syntax. Each layer is specified by `.` and the outside must be quoted with <code>\`\`</code> to notify that it is a jq syntax. For example , given a nested field
```json
{
    "properties": {
        "nested-layer": {
            "some-field": "some-value"
        }
    }
}
```
the schema should be
```json
{
    "rules": {
        "`.nested-layer.some-field`": {
        }
    }
}
```
