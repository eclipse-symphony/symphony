# Property expressions

Symphony solution [components](./solution.md#componentspec) and campaign stages can use property expressions in their property values, wrapped in `${{}}`. Property expressions are evaluated at the solution [vendor](../vendors/overview.md) level, which means that they are evaluated immediately once they arrive at the Symphony API surface. Hence, none of the [managers](../managers/overview.md) or [providers](../providers/_overview.md) need to worry about (or be aware of) the expression rules. When authoring Symphony artifacts, a user can use these expressions in their solution documents.

In general, Symphony attempts to parse the property values as strings whenever possible. Only when Symphony detects clear and valid arithmetical expressions and function calls will it try to evaluate them first before the string evaluation. This means that Symphony is mostly tolerable to syntax errors in expressions and will treat those as string literals. For example, `"10/2"` is interpreted as a division, while `"10/0"` is treated as a string because otherwise it's a divide-by-zero error.

## Constants

* Integer constants are evaluated as numbers when possible.
* Float numbers are always treated as strings. However, if a float is the result of evaluation (such as `5/2`), it’s treated as number until no longer possible. Hence `5/2+1` is `3.5` while `5/2+a` is `2.5a`.
* Negative integers are allowed.
* String constants can't have spaces between words. Surrounding spaces are trimmed. Words after the first space are omitted. To use spaces between words, surround the string literal with single quote(`'`).
* For single-quoted strings, single quotes are removed. For double-quoted strings, double quotes are kept as they are.

The following table summarizes how constants are evaluated:

| Expression | Value | Comment |
|--------|--------|--------|
| `"1"` | `"1"` | Integers are treated as strings when no further evaluation is possible |
| `"3.14"` | `"3.14"` | Floats are always treated as strings |
| `" "` | `""` | Spaces are trimmed |
| `"  abc   "` | `"abc"`| Spaces are trimmed |
| `"abc def"`| `"abc"`| Words after first space are omitted|
| `'abc def'`|`"abc def"` | Single-quoted string|
| `'abc def`|`"'abc def"` | Single-quote not closed, returns as it is|
| `'abc def'hij`| `"abc def"`| Single-quoted part is parsed, the rest is omitted|

## Operators

* Plus (`+`) is treated as numeric addition, if both sides of `+` evaluate to numbers. Otherwise, it’s treated as string concatenation.
* Minus (`-`) is treated as numeric addition, if both sides of `-` evaluate to numbers. Otherwise, it’s treated as a dash (`-`).
* Multiplication (`*`) is treated as numeric multiply, if both sides of `*` evaluate to numbers. Otherwise, it’s treated as a star (`*`).
* Divide (`/`) is treated as numeric addition, if both sides of `/` evaluate to numbers (divide by zero is not allowed). Otherwise, it’s treated as a forward-slash (`/`).
* Use parenthesis (`()`) to change calculation precedence. Expressions in parenthesis are evaluated first.

The following table summarizes how operators work:

| Expression | Value | Comment |
|--------|--------|--------|
| `"+2"`| `"2"`| Unary plus|
| `"-3"`| `"-2"`| Unary minus|
| `"-abc"`| `"-abc"`| Dash ABC |
| `"abc-"`| `"abc-"`| ABC dash |
| `"+a"` | `"a"`| (empty) plus a |
| `"a+"` | `"a"`| a plus (empty) |
| `"1+2"`| `"3"`| Addition (integers are treated as numbers when possible)|
| `"1.0+2.0"`|`"1.02.0"`| Concatenation (floats are always treated as strings)|
| `"1-2"`|`"-1"`| Subtraction |
| `"3*-4""`|`"-12"`| Multiplication|
| `"3-(1+2)/(2+1)"`|`"2"`| Parentheses are evaluated first |

> **NOTE:** **Why do we treat integers as numbers when possible, and floats always as strings?** We allow integer calculations for cases like calculating an offset and other simple number manipulations. However, we don’t aim to offer a full-scale scripting language, and floats in our contexts are often version numbers like `“1.2”` and `“1.2.3”`. Hence we always treat floats as strings. Of course, you can do things like `“1.(3+4)”`, which evaluates to `“1.7”` because numerical evaluation of the integer expression `(3+4)` is still possible.

## Functions

Symphony supports a few functions for artifact content overrides, as summarized by the following table.

When these functions are used, a valid `EvaluationContext` is required, which injects a secret provider, a configuration provider, and a Symphony deployment object as evaluation context. Failing to provide required context causes an evaluation error.

| Function | Behavior|
|----------|---------|
|`$config(<config object>, <config key>, [<overrides>])` | Reads a configuration from a config provider |
|`$context([<JsonPath>])` | Reads the evaluation context value. If a JsonPath is specified, it applies the path to the context value (same as `$val()`) |
|`$input(<field>)` | Reads campaign activation input `<field>` |
|`$instance()`| Gets instance name of the current deployment |
|`$json(<value>)`| Arranges `<value>` into a JSON string |
|`$output(<stage>, <field>)` | Reads the output `<field>` value from a campaign `<stage>` outputs|
|`$param(<parameter name>)`| Reads a component parameter. Parameters are defined on [component](./solution.md#componentspec) and can be overridden by stage arguments in [instance](./instance.md). |
|`$property(<property name>)`| Reads a property from the evaluation context |
|`$secret(<secret object>, <secret key>)`| Reads a secret from a secret store provider |
|`$val([<JsonPath>])` | Reads the evaluation context value. If a JsonPath is specified, it applies the path to the context value (same as `$context()`) |

Symphony also supports common logical operators:

| Function | Behavior|
|----------|---------|
|`$and(<condition1>, <condition2>)` | `true` if both conditions evaluate to `true` (boolean) or `"true"` (string)|
|`$between(<value>, <value1>,<value2>)` | `true` if `<value>` is between `<value1>` and `<value2>` |
|`$equal(<value1>, <value2>)` | `true` if `<value1>` equals `<value2>` |
|`$ge(<value1>, <value2>)` | `true` if `<value1>` is greater or equal to `<value2>` |
|`$gt(<value1>, <value2>)` | `true` if `<value1>` is greater than `<value2>` |
|`$if(<condition>)` | `true` if `<condition>` evaluates to `true` (boolean) or `"true"` (string)|
|`$in(<value>, [<ref-value>])` | `true` if `<value>` exists in the list of `<ref-value>` |
|`$le(<value1>, <value2>)` | `true` if `<value1>` is less or equal to `<value2>` |
|`$lt(<value1>, <value2>)` | `true` if `<value1>` is less than `<value2>` |
|`$not(<condition>)` | `true` if `<condition>` evaluates to `false` (boolean) or `"false"` (string)|
|`$or(<condition1>, <condition2>)` | `true` if either `<condition1>` or `<condition2>` evaluates to `true` (boolean) or `"true"` (string)|

## Evaluation context

Functions like `$input()`, `$output()`, `instance()`, `property()` and  `$val()` etc. can be only evaluated in an appropriate evaluation context, to which Symphony automatically injects contextual information, such as Campaign activation inputs. When you use Symphony API, the evaluation context is automatically managed so you can use these functions in appropriate contexts without concerns. However, using these functions outside of an appropriate context leads to an error.

## Use operators as characters

We try to parse properties as closely as strings as possible with limited calculations and functions calls allowed. When operators are used out of the context of an expression, they are evaluated differently. Although the following are unlikely scenarios, we present how they are evaluated following the above evaluation rules.

* When used alone, a period (`.`) is returned as it is, such as `.` and `...`.
* When used alone, a plus(`+`) is treated as a unary operator, which means one or more `+` signs are evaluated to empty strings, as “plus nothing” is still “nothing”.
* When used alone, a minus(`-`) is treated as a unary operator, which means a single `-` is evaluated to empty string, as “minus nothing” is “nothing”. However, when you use two minus signs, the second minus is treated as “dash”, and a negative “dash” is still “dash”. Hence `--` evaluates to `-`.
* When used alone, a forward-slash (`/`) is returned as it is, such as `/` and `///`.

## Skip parsing

The parser can't parse arbitrary strings as an expression. For example, complex file paths and URLs (with parameters and encodings) are probably parsed incorrectly. To skip the expression parsing, you need to put the string into a pair of single quotes (`'`), for example:

```json
"'c:\\demo\\HomeHub.Package_1.0.9.0_Debug_Test\\HomeHub.Package_1.0.9.0_x64_Debug.appxbundle'"
```

and

```json
"'https://manual-approval.azurewebsites.net:443/api/approval/triggers/manual/invoke?api-version=2022-05-01&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=<secret>'"
```

In such cases, the string is returned as it is (with surrounding single quotes removed). You can also partially skip parsing by using string concatenations. For example:

```bash
parsed + 'this is not parsed'
```
