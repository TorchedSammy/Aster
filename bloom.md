# Bloom Language Specification
Bloom is a DSL created to facilitate the Aster shell. It was
designed to be simple and easy enough to parse and read and use
in a shell. This means Bloom has a few amount of features
and syntax.

In this document, `->` at the beginning of a line
in a code block is an indicator for a REPL line.

## Commands
A command is a space separated list where the first element is the
command name. Commands are one of the center points of Bloom.

```sh
print "hello"
```

### Builtin Commands
#### print
The `print` command prints the passed arguments,

Example:
```go
-> print "hello"
// Prints hello
```

## Strings
Anything enclosed in double or single quotes are strings.
Some examples:
- `"hello world"`
- `'bloom'`
- `"milk \x2b sugar"` (it's `milk + sugar`, by the way)

Bloom strings support escaping characters with a backslash (`\`)
and there is currently 1 supported escape sequence.

- `\xNN`: This sequence encodes NN into the equivalent ASCII
value

## Numbers
Numbers are digits only, without support for decimals.
Example: `420`

## Variables
Variables can be declared *and* assigned with the `var` keyword,
as so:
```go
var hello = ""
```

### References
Variables can be referenced with the `#` operator and used as a literal
value.

```lua
var who = "aster"
print "Hello " #hello
```
