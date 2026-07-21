---
type: docs
title: "Symphony Expression"
linkTitle: "Symphony Expression"
description: ""
weight: 60
---

Symphony Expression is the template and evaluation syntax used by Symphony runtime components to compute values dynamically.

## Syntax

Expressions are wrapped with `${{ ... }}`.

Examples:

- Literal expression: `${{1}}`
- Arithmetic expression: `${{1+2}}`
- Function call: `${{$if($gt(3,2), ok, fail)}}`
- Mixed text and expressions: `http://abc.com:${{8080+1}}/api`

## Evaluation Rules

- A plain string without `${{...}}` is returned as-is.
- A single expression can return numbers, booleans, strings, maps, or arrays.
- Multiple segments are concatenated as strings.
- Empty expression `${{ }}` evaluates to null.

## Operators

Supported operators in expressions:

- Arithmetic: `+`, `-`, `*`, `/`
- Concatenation-style tokens: `.`, `:`, `?`, `=`, `&`, `~`, `\\`
- Grouping: `(...)`

Notes from runtime behavior:

- Integer arithmetic returns numbers, for example `${{1+2}} -> 3`.
- Float-looking literals are often treated as version-like strings, for example `${{6.3}} -> "6.3"`.
- Computed float results are numeric, for example `${{1/2}} -> 0.5`.
- Division by zero returns string-form output, for example `${{10/0}} -> "10/0"`.

## Built-in Functions

| Function | Signature | Description |
|--------|--------|--------|
| `$param` | `$param(key)` | Reads a component parameter from deployment context. |
| `$property` | `$property(key)` | Reads a property from context properties. |
| `$input` | `$input(key)` | Reads a key from stage inputs. |
| `$output` | `$output(step, key)` | Reads an output value from a named step. |
| `$trigger` | `$trigger(key, default)` | Reads a trigger value; falls back to default when missing. |
| `$equal` | `$equal(a, b)` | Equality comparison. |
| `$and` | `$and(a, b)` | Boolean AND. |
| `$or` | `$or(a, b)` | Boolean OR. |
| `$not` | `$not(a)` | Boolean NOT. |
| `$gt` | `$gt(a, b)` | Greater-than numeric comparison. |
| `$ge` | `$ge(a, b)` | Greater-than-or-equal numeric comparison. |
| `$lt` | `$lt(a, b)` | Less-than numeric comparison. |
| `$le` | `$le(a, b)` | Less-than-or-equal numeric comparison. |
| `$between` | `$between(v, min, max)` | Checks numeric range inclusively. |
| `$if` | `$if(condition, whenTrue, whenFalse)` | Conditional expression. |
| `$in` | `$in(value, option1, option2, ...)` | Membership check. |
| `$config` | `$config(object, field, overlays...)` | Reads from config provider. |
| `$secret` | `$secret(object, field)` | Reads from secret provider. |
| `$instance` | `$instance()` | Returns current deployment instance name. |
| `$val` / `$context` | `$val()` or `$val(path)` | Returns current context value or extracts a path. |
| `$base64encode` | `$base64encode(value)` | Base64 encodes a value as string. |
| `$base64decode` | `$base64decode(value)` | Base64 decodes a string value. |
| `$jsonpath` | `$jsonpath(value, path)` | Applies JSONPath query to JSON-like value. |
| `$json` | `$json(value)` | Serializes value to JSON string. |
| `$str` | `$str(value)` | Converts value to string. |

## Context Requirements

Some functions require runtime context providers or objects:

- `$config` requires a config provider.
- `$secret` requires a secret provider.
- `$param` and `$instance` require deployment context.
- `$output`, `$input`, `$trigger`, `$property` require corresponding collections.

When required context is missing, evaluation returns an error.

## Common Examples

- Conditional stage selection:
  - `${{$if($lt($output(test,foo), 5), nextStage, '')}}`
- Read deployment instance name:
  - `${{$instance()}}`
- Read a property and compare:
  - `${{$equal($property(OS), windows)}}`
- Query nested value using JSONPath:
  - `${{$jsonpath($val(), '$.status')}}`

## Notes

This reference reflects runtime behavior implemented in parser and utility code. Edge-case parsing behavior (for example multiple identifiers inside one expression) follows current tests and may evolve over time.
