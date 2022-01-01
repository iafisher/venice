# The Venice Development Guide
This is a manual for working on the Venice codebase. Most Venice programmers will never need to read this manual. If you are interested in *learning* Venice, you should read the [tutorial](https://github.com/iafisher/venice/blob/master/docs/tutorial.md) instead.

This document should be read in conjunction with the [contributing guide](https://github.com/iafisher/venice/blob/master/CONTRIBUTING.md), which concerns the logistics of contributing to the open-source project.


## Getting started
Venice is written in the [Go programming language](https://go.dev/) and hosted on [GitHub](https://github.com/iafisher/venice).

To compile Venice, you will need to install [Go 1.16](https://go.dev/learn/) and [git](https://github.com/git-guides/install-git). (Older versions of Go may or may not work.)

Once Go and git are installed, you can clone and build the project with:

```
git clone https://github.com/iafisher/venice
cd venice
go build
```

This will produce a `venice` binary in your local directory.


## Architecture
Venice code compiles to [bytecode](https://en.wikipedia.org/wiki/Bytecode) which runs on a [virtual machine](https://en.wikipedia.org/wiki/Virtual_machine#Process_virtual_machines). Thus, the Venice language implementation logically comprises two programs: a compiler which translates Venice source code into Venice bytecode (analogous to the `javac` binary for Java), and a virtual machine which executes Venice bytecode (analogous to the `java` binary for Java).

For simplicity of use, the present Venice implementation is a single program that combines the compiler and the virtual machine, as well as an interactive read-eval-print loop (REPL).

A Venice program goes through four stages during compilation:

- Lexing: The program is tokenized, i.e. converted from a string (`1 + 1`) into a stream of tokens (`1`, `+`, `1`)
- Parsing: The program is parsed into an abstract syntax tree.
- Compilation: The abstract syntax tree is checked for type errors and compiled into bytecode.
- Execution: The bytecode is executed by a virtual machine.

The first three steps happen in the compiler, and the last step happens in the virtual machine.


## Layout
Besides the main entry point at `main.go`, all of the source code lives under the `src/` directory, which contains four packages:

- `common/`: Common code shared by the compiler and the virtual machine.
- `compiler/`: The compiler, which translates Venice source code into Venice bytecode.
- `test/`: End-to-end tests which evaluate Venice code snippets.
- `vm/`: The virtual machine, which executes Venice bytecode.

The `compiler` and `vm` packages should not depend on each other; any code that is needed by both should be placed in the `common` package instead.

The `common` package:

- `src/common/bytecode`: The definition of the Venice bytecode format.
- `src/common/lex`: The lexer for Venice. This is in `common` because the virtual machine's bytecode parser reuses the same lexer as the compiler for simplicity.

The `compiler` package:

- `src/compiler/ast.go`: The definitions of Venice abstract syntax trees (ASTs).
- `src/compiler/compiler.go`: The main component of the compiler, which translates abstract syntax trees into bytecodes and performs type-checking.
- `src/compiler/parser.go`: The Venice parser.
- `src/symbol_table.go`: A symbol table type for the compiler, and type declarations for built-in functions.
- `src/vtype.go`: The definitions of Venice types (not to be confused with the definitions of Venice *objects* in the `vm` package).

The `vm` package:

- `src/vm/builtins.go`: Implementations of built-in functions.
- `src/vm/vm.go`: The main component of the virtual machine, which executes bytecode.
- `src/vm/vval.go`: The definitions of Venice objects (not to be confused with the definitions of Venice *types* in the `compiler` package).


## Common workflows
### Adding a new built-in function
To add a new built-in function (either a global built-in, like the `print` function, or a method on a built-in object, like `list.slice`), you need to:

1. Add an entry to the compiler's symbol table, in `src/compiler/symbol_table.go`. Global built-ins are defined in `NewBuiltinSymbolTable`. Built-in object methods are defined in various map objects, e.g. `listBuiltins`, `mapBuiltins`, etc.
2. Implement the method in `src/vm/builtins.go`. Make sure to also add an entry in the `builtins` map at the bottom of the file.

[This commit](https://github.com/iafisher/venice/commit/cdef1ed4304f7fab7c93ca0b3454c806903ed7c3) is a good example of writing a simple built-in function.


## Testing
Some components, notably the lexer and the parser, have unit tests located in `_test.go` files (e.g., `src/compiler/parser_test.go`). Otherwise, Venice is tested with two mechanisms:

- Internal end-to-end tests in the `src/test` package, which test small snippets of Venice code.
- External end-to-end tests in the top-level `tests/` directory, which are run with `./test_e2e` and execute larger Venice programs from disk.

Internal end-to-end tests are required for all new language features, and are generally preferred to external end-to-end tests.


## Style
Venice is formatted with `go fmt`. The soft line limit is 90 characters, but lines may reach 95 characters or so if splitting them into multiple lines would be awkward.

### Naming
The Venice codebase prefers more verbose names compared to typical Go code. Take a look at existing code in the file you are editing, and try to follow local conventions. For example, AST variables in `src/compiler/compiler.go` tend to be called `node`.

One common idiom is to name a variable of an interface type as `fooAny` and its concrete type as `foo`, for example:

```
veniceObjectAny := args[0]
switch veniceObject := veniceObjectAny.(type) {
	...
}
```

### Struct initializers
Always include the names of fields in struct initializers, unless the struct only has one field. For example:

```
&VeniceMapType{
	KeyType: keyType,
	ValueType: valueType,
}
```
