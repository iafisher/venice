# Venice
Venice is a modern, high-level, statically-typed programming language. It pairs the convenience of Python with the static typing and modern language features of Rust.

<!-- TODO(2021-07-21): Put a concise but demonstrative example of Venice code here. -->

For a comprehensive description of the language, see `docs/reference.md`.


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

Each of these four steps corresponds to a package under `src/`:

- `lexer`
- `parser`
- `compiler`
- `vm`

A few other packages define critical data structures:

- `ast`: Abstract syntax trees
- `bytecode`: Venice bytecode instructions
- `vtype`: Representation of Venice types (e.g., `int`, `string`, `list<int>`)
- `vval` Representation of Venice objects (e.g., `42`, `"hello"`, `[1, 2, 3]`)

Finally, there are a couple of miscellaneous packages:

- `exec`: Contains most of the project's end-to-end tests.
- `wasm`: A proof-of-concept for executing Venice code in a browser using Web Assembly.
