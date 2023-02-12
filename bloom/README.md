![](https://safe.kashima.moe/24bk9ki9km3v.png)

Bloom is a simple and fast DSL for defining image filters and running
commands. Although it is made for Aster use specifically, the hope is
that it can still be used for other basic scripting tasks and other
image-related utilities in Go or else.

Bloom is designed to be easy and simple to parse, read, and write
in a shell context. This means there is a small set of features and
syntax.

Some example code:  
```go
// a filter runs through an image and is automatically passed
// the current working image
filter sepia {
	// op sets certain operators about the filter
	// this sets the operator `filterType` to `color`.
	// which means that this filter will alter the colors of the image
	op filterType 'color'

	var r = (#r * 0.393) + (#g * 0.769) + (#b * 0.189)
	var g = (#r * 0.349) + (#g * 0.686) + (#b * 0.168)
	var b = (#r * 0.272) + (#g * 0.534) + (#b * 0.131)

	return #r, #g, #b
}

// meanwhile a function could do anything
function hello(person) {
	if #person == "" {
		var person = "world"
	}

	print "Hello" .. #person .. "!"
}
```

## Bloom Language Specification
## Commands
A command is a space separated list where the first element is the
command name. Commands are one of the center points of Bloom.

```sh
print "hello"
```

### Semicolons
A semicolon can be used to split up commands on a single line:
```sh
foo; bar
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
print "Hello " #who
```

## Filter Declarations
Filter Declarations are the 2nd center point of Bloom. A filter passes
through an image and alters it, for example a Sepia filter. A Filter
Declaration defines these filters.

A filter acts the same as a command from outside use (like in the Aster
shell) but is ran at every pixel of an image. Depending on the type
of filter it will have certain globals that is expected to be modified
and returned.

```go
filter name {
	op filterType 'color'

	var r = #r * 2
	var g = #g * 2
	var b = #b * 2

	return #r, #g, #b
}
```

### Filter Types
Currently there is only 1 filter type, which is `color`.
A color filter alters R, G, B values of an image.
