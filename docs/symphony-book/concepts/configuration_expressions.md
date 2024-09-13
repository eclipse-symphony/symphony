# Configuration expressions

The syntax of the `$config` expression is

```
$config(<config object>, <config key>, [<overrides>])
```

When provided with a `<config key>`, the configuration attempts to parse and evaluate the `<config key>` from the `<config object>`. The parsing process of `<config key>` leverages the [gojq](https://github.com/itchyny/gojq?tab=readme-ov-file) library, which supports most of the jq syntax such as select, slice, filter, map and transform. Some exceptional cases of gojq are listed in [difference to jq](https://github.com/itchyny/gojq?tab=readme-ov-file#difference-to-jq).

Details about overrides can be found in [inheritance](../configuration-management/inheritance.md) and is not discussed here.

## Examples of usage

> **Note**
`$config` expressions support both direct field reference as well as jq syntax. To use jq syntax inside `$config` expressions, make sure to quote the query with <code>``</code> outside the query. Otherwise, it will be evaluated as a simple field reference.

The following examples will utilize below configuration as reference.
```yaml
spec:
  rootResource: config1
  catalogType: config
  properties:
    myKey:
      subVersion: 1
      subList: [
        "Tag1",
        "Tag2"
      ]
    dotted.key: dottedValue
    
```

- Access direct fields inside properties. Fields can be accessed directly, without `` quote:
```yaml
$config(config1, myKey)
# result: {subKey: 1, subList: ["Tag1", "Tag2"]}
```

Must quote with `` to access nested fields or perform complex operations in queries:
- Access nested keys:
```yaml
$config(config1, `.myKey.subKey`)
# result: 1
$config(config1, `.myKey.subList`)
# result: ["Tag1","Tag2"]
```

- Slicing:
```yaml
$config(config1, `.myKey.subList[1:2]`)
# result: ["Tag2"]
```

- Iterator (**different with gojq**):
`jq` has the iterator syntax (such as `[]`) which will iterate through all the qualified result and return one at a time. `$config` will return the last qualified result from the iterator. 
```yaml
$config(config1, `.myKey.subList[]`)
# result: ["Tag2"]
```

- Array construction
```yaml
$config(config1, `[.myKey.subList[]]`)
# result: ["Tag1","Tag2"]
```

- Comparisons:
```yaml
$config(config1, `.myKey.subVersion < 2`)
# result: true
```

- Pipe:
```yaml
$config(config1, `.myKey | .subVersion`)
# result: 1
```

- Regular expression:
```yaml
$config(config1, `.myKey.subList[0] | test("tag+")`)
# result: true
```

More usage of jq syntax can be found at [jq Manual](https://jqlang.github.io/jq/manual/).

## Quotation marks

In `jq`, the double quote is used to specify an internal unit of expression, such as a regular expression, or a specific name of a key. In these situations, the back quote <code>`</code> must be used outside the expression to be distinguished with the usage of double quote. In addition, to notify symphony to skip parsing the expression, you need to add the single quote outside the entire <config key> expression.

For example, to correctly access the value of `dotted.key`, the correct syntax should be
```yaml
$config('config1','`.\"dotted.key\"`')
# result: dottedValue
```
