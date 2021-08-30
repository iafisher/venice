# The Venice programming language
Venice is a modern, high-level, statically-typed programming language. It pairs the convenience of Python with the static typing and modern language features of Rust.

<!-- Venice's syntax is closest to Rust's, so we use that as the syntax declaration for the code block. -->
```rust
import map, join from "itertools"

enum Json {
  JsonObject({string: Json}),
  JsonArray([Json]),
  JsonString(string),
  JsonNumber(real),
  JsonBoolean(bool),
  JsonNull,
}

fn serialize_json(j: Json) -> string {
  match j {
    case JsonObject(obj) {
      let it = "${key}: ${serialize_json(value)}" for key, value in obj
      return "{" ++ join(it, ", ") ++ "}"
    }
    case JsonArray(values) {
      return "[" ++ join(map(values, serialize_json), ", ") ++ "]"
    }
    case JsonString(s) {
      return s.quoted()
    }
    case JsonNumber(x) {
      return string(x)
    }
    case JsonBoolean(x) {
      return string(x)
    }
    case JsonNull {
      return "null"
    }
  }
}
```

For a full introduction to the language, read the [tutorial](https://github.com/iafisher/venice/blob/master/docs/tutorial.md).

**NOTE**: Venice is in the early stages of development, and not yet ready for production use.


## Installation
Install Venice with Go:

```
$ go get github.com/iafisher/venice
```

The Venice binary will be installed at `$GOBIN/venice`.

Run `venice` to open the interactive read-eval-print loop (REPL), `venice compile example.vn` to compile a Venice program to bytecode, or `venice execute example.vn` to compile and execute a program in one step.


## Development
Venice is written in the Go programming language. Besides the entry point at `main.go`, all the source code lives in the `src/` directory.

A Venice program goes through four stages during compilation:

- Lexing: The program is tokenized, i.e. converted from a string (`1 + 1`) into a stream of tokens (`1`, `+`, `1`)
- Parsing: The program is parsed into an abstract syntax tree.
- Compilation: The abstract syntax tree is checked for type errors and compiled into bytecode.
- Execution: The bytecode is executed by a virtual machine.

The first three stages correspond to the packages under `src/compiler`: `src/compiler/lexer`, `src/compiler/parser`, and `src/compiler/compiler`. The last stage corresponds to the `src/vm` package.

A few other packages and files define critical data structures:

- `compiler/ast`: Abstract syntax trees
- `common/bytecode`: Venice bytecode instructions
- `compiler/vtype`: Representation of Venice types (e.g., `int`, `string`, `[int]`)
- `vm/vval.go` Representation of Venice objects (e.g., `42`, `"hello"`, `[1, 2, 3]`)
