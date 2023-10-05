# Property Expressions

Symphony Solution [```Components```](./solution.md#componentspec) and Campaign Stages can use property expressions in their property values. Property expressions are evaluated at Solution [Vendor](../vendors/overview.md) level, which means they are immediately evaluated once they arrive at the Symphony API surface. Hence, none of the [Managers](../managers/overview.md) or [Providers](../providers/overview.md) need to worry about (or be aware of) the expression rules. When authoring Symphony artifacts, a user can use these expressions in their Solution documents.

In general, Symphony attempts to pare the property values as close as strings as possible. Only when Symphony detects clear and valid arithmetical expressions and function calls, it will try to evaluate them first before the string evaluation. This means Symphony is mostly tolerable to syntax errors in expressions and will treat those as string literals, unless there are clear errors such trying to divide by zero. For example, ```"10/"``` is allowed as Symphony assumes ```/``` is used as a forward-slash. However, ```"10/0"``` is disallowed as it leads to division by zero. ```"'10/0'"``` is allowed as it’s a quoted string.

## Constants

* Integer constants are evaluated as numbers when possible.
* Float numbers are always treated as strings. However, if a float is the result of evaluation (such as ```5/2```), it’s treated as number till no longer possible. Hence ```5/2+1``` is ```3.5``` while ```5/2+a``` is ```2.5a```.
* Negative integers are allowed.
* String constants can't have spaces between words. Surrounding spaces are trimmed. Words after the first space are omitted. To use spaces between words, surround the string literal with single quote(```'```).
* For single-quoted strings, single quotes are removed. For double-quoted strings, double quotes are kept as they are.


The following table summarizes how constants are evaluated:

| Expression | Value | Comment |
|--------|--------|--------|
| ```"1"``` | ```"1"``` | Integers are treated as strings when no further evaluation is possible |
| ```"3.14"``` | ```"3.14"``` | Floats are always treated as strings
| ```" "``` | ```""``` | Spaces are trimmed |
| ```"  abc   "``` | ```"abc"```| Spaces are trimmed |
| ```"abc def"```| ```"abc"```| Words after first space are omitted|
| ```'abc def'```|```"abc def"``` | Single-quoted string|
| ```'abc def```|```"'abc def"``` | Single-quote not closed, returns as it is|
| ```'abc def'hij```| ```"abc def"```| Single-quoted part is parsed, the rest is omitted|

## Operators
* Plus(```+```) is treated as numeric addition, if both sides of ```+``` evaluate to numbers. Otherwise, it’s treated as string concatenation. 
* Minus (```-```) is treated as numeric addition, if both sides of ```-``` evaluate to numbers. Otherwise, it’s treated as a dash (```-```). 
* Multiplication (```*```) is treated as numeric multiply, if both sides of ```-``` evaluate to numbers. If the left of * is a string and the right of * is a number, the string is repeated the number of times. Otherwise, it’s treated as a star (```*```). 
* Divide (```-```) is treated as numeric addition, if both sides of ```-``` evaluate to numbers (divide by zero is not allowed). Otherwise, it’s treated as a forward-slash (```/```). 
* Use parenthesis (```()```) to change calculation precedence. Expressions in parenthesis are evaluated first.

The following table summarizes how operators work:
| Expression | Value | Comment |
|--------|--------|--------|
| ```"+2"```| ```"2"```| Unary plus|
| ```"-3"```| ```"-2"```| Unary minus|
| ```"-abc"```| ```"-abc"```| Dash ABC |
| ```"abc-"```| ```"abc-"```| ABC dash |
| ```"+a"``` | ```"a"```| (empty) plus a |
| ```"a+"``` | ```"a"```| a plus (empty) |
| ```"1+2"```| ```"3"```| Addition (integers are treated as numbers when possible)|
| ```"1.0+2.0"```|```"1.02.0"```| Concatenation (floats are always treated as strings)|
| ```"1-2"```|```"-1"```| Subtraction |
|  ```"3*-4""```|```"-12"```| Multiplication|
| ```"abc*3"```|```"abcabcabc"```| Repeat abc 3 times |
| ```"abc*-3"```|```"abc*-3"```| Repeating -3 times is impossible, return as a string|
| ```"abc*0"```| ```""```| Repeat 0 times |
| ```"abc*(5/2)"```| ```"abcabc"```| Repeat floor(5/2) times |
|```"3-(1+2)/(2+1)"```|```"2"```| Parentheses are evaluated first |


> **NOTE:** **Why do we treat integers as numbers when possible, and floats always as strings?** We allow integer calculations for cases like calculating an offset and other simple number manipulations. However, we don’t aim to offer a full-scale scripting language, and floats in our contexts are often version numbers like “```1.2```” and “```1.2.3```”. Hence we always treat floats as strings. Of course, you can do things like “```1.(3+4)```”, which evaluates to “```1.7```” because at ```(3+4)``` numerical evaluation of integer expression is still possible.
## Functions
Symphony supports a few functions for artifact content overrides, as summarized by the following table.

When these functions are used, a valid ```EvaluationContext``` is required, which injects a secret provider, a configuration provider, and a Symphony deployment object as evaluation context. Failing to provide required context causes an evaluation error.

| Function | Behavior|
|--------|--------|
|```$and(<condition1>, <condition2>)``` | If both ```<condition1>``` and ```<condition2>``` evaluate to ```true``` (boolean) or ```"true"``` (string)|
|```$between(<value>, <value1>,<value2>)``` | If ```<value>``` is between ```<value1>``` and ```<value2>``` |
|```$config(<config object>, <config key>, [<overrides>])``` | Reads a configuration from a config provider |
|```$equal(<value1>, <value2>)``` | if ```<value1>``` equals to ```<value2>``` |
|```$ge(<value1>, <value2>)``` | if ```<value1>``` is greater or euqal to ```<value2>``` |
|```$gt(<value1>, <value2>)``` | if ```<value1>``` is greater than ```<value2>``` |
|```$if(<condition>)``` | If ```<condition>``` evaluates to ```true``` (boolean) or ```"true"``` (string)|
|```$in(<value>, [<ref-value>])``` | If ```<value>``` exists in the list of ```<ref-value>``` |
|```$input(<field>)``` | Reads Campaign activation input ```<field>``` |
|```$instance()```| Gets instance name of the current deployment |
|```$le(<value1>, <value2>)``` | if ```<value1>``` is less or euqal to ```<value2>``` |
|```$lt(<value1>, <value2>)``` | if ```<value1>``` is less than ```<value2>``` |
|```$not(<condition>)``` | If ```<condition>``` evaluates to ```false``` (boolean) or ```"false"``` (string)|
|```$output(<stage>, <field>)``` | Reads the output ```<field>``` value from a Campaign ```<stage>``` outputs|
|```$or(<condition1>, <condition2>)``` | If either ```<condition1>``` or ```<condition2>``` evaluates to ```true``` (boolean) or ```"true"``` (string)|
|```$param(<parameter name>)```| Reads a component parameter<sup>1</sup>|
|```$property(<proerty name>)```| Reads a property from the evaluation context |
|```$secret(<secret object>, <secret key>)```| Reads a secret from a secret store provider |
|```$val([<JsonPath>])``` | Reads the evaulation context value. if a JsonPath is specified, apply the path to the context value |

<sup>1</sup>: Parameters are defined on [Component](./solution.md#componentspec) and can be overridden by stage Arguments in [Instance](./instance.md).

## Evaluation Context
Functions like ```$input()```, ```$output()```, ```instance()```, ```property()``` and  ```$val()``` etc. can be only evaluated in an appropriate evaluation context, to which Symphony automatically injects contextual information, such as Campaign activation inputs. When you use Symphony API, the evaluation context is automatically managed so you can use these functions in appropriate contexts without concerns. However, using these functions outside of an appropriate context leads to an error.

## Using Operators Alone

We try to parse properties as closely as strings as possible with limited calculations and functions calls allowed. When operators are used out of the context of an expression, they are evaluated differently. Although the following are unlikely scenarios, we present how they are evaluated following the above evaluation rules.

* When used alone, a period (```.```) are returned as it is, such as ```.``` and ```...``` are returned as they are.
* When used alone, a plus(```+```) is treated as a unary operator, which means a single, or a consecutive ```+``` signs are evaluated to empty strings, as “plus nothing” is still “nothing”.
* When used alone, a minus(```-```) is treated as a unary operator, which means a single ```-``` is evaluated to empty string, as “minus nothing” is “nothing”. However, when you use two minus signs, the second minus is treated as “dash”, and a negative “dash” is still “dash”. Hence ```--``` evaluates to ```-```. 
* When used alone, a forward-slash (```/```) are returned as it is, such as ```/``` and ```///``` are returned as they are.

## Skipping Parsing
The parser can't parse arbitrary strings as an expression. For example, complex file paths and URLs (with parameters and encodings) are probably parsed incorrectly. To skip the expression parsing, you need to put the string into a pair of single quotes (```'```), for example:
```json
"'c:\\demo\\HomeHub.Package_1.0.9.0_Debug_Test\\HomeHub.Package_1.0.9.0_x64_Debug.appxbundle'"
```
and 
```json
"'https://manual-approval.azurewebsites.net:443/api/approval/triggers/manual/invoke?api-version=2022-05-01&sp=%2Ftriggers%2Fmanual%2Frun&sv=1.0&sig=<secret>'"
```
In such cases, the string is returned as it is (with surrounding single quotes removed). You can also partially skip parsing by using string concatations. For example:
```bash
parsed + 'this is not parsed'
```
