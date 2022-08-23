# The Venice implementation
WORK IN PROGRESS

## Introduction
This document describes the design of the compiler for the Venice programming language. For a description of the language itself, see the [language reference](https://github.com/iafisher/venice/blob/master/docs/language-reference.md).

## Goals
The focus of the first version of the compiler, which will be written in Rust, is correctness and simplicity. Efficiency, either in the sense of emitting efficient code or compiling programs quickly, is not a focus. Error messages for incorrect programs should give enough information for the programmer to be able to identify the error (i.e., the location in the source code and a clear description of the problem), but they do not need to be as detailed as, e.g., Rust's.

These goals may change in future versions of the compiler.

## Compiler architecture
The fundamental task of the compiler is to translate a Venice program written in textual source code into an executable file in x86-64 assembly code.

For simplicity, the first version of the compiler will emit a textual assembly-language program which an external assembler and linker can turn into an actual binary executable. Future versions of the compiler will directly produce ELF binaries.

Translation occurs in three phases:

- Lexical and syntactic analysis
- Semantic analysis
- Code generation

...and involves three major data structures:

- Parse tree
- Abstract syntax tree
- Control-flow graph

The output of the lexical and syntactic analysis phase is a concrete parse tree whose nodes correspond directly to textual elements in the program's source code. For example, if there is a `while` loop in the source code, there will be a corresponding `while` node in the parse tree.

The parse tree is converted into an abstract syntax tree, which has a similar structure but simplifies some syntactic forms, and adds type information to each node. For example, `if` statements in Venice can have an arbitrary number of `else if` clauses, but in the abstract syntax tree there are only `if-else` nodes and `else if` clauses are represented by nesting two or more `if-else` nodes.

The abstract syntax tree is lowered into a control-flow graph representation using an abstract assembly language similar to LLVM's intermediate representation. Some aspects of the assembly language are idealized, e.g. there are an infinite number of registers available, and function callers don't have to explicitly place arguments in registers or on the stack.

In the last step, the control-flow graph is converted into concrete machine code for a specific architecture (in Venice's case, x86-64).

<https://mapping-high-level-constructs-to-llvm-ir.readthedocs.io/en/latest/README.html>

## Lexical and syntactic analysis
### The lexer
The lexer is responsible for breaking up the text of the program into tokens such as symbols, numbers, and punctuation.

```
class Lexer:
  def __init__(self, program: str) -> None:
    ...

  def token(self) -> Token:
    ...

  def advance(self) -> None:
    ...
```

`Lexer.token` returns the current token in the program, and `Lexer.advance` advances to the next token. If the end of the program has been reached, then `Lexer.token` will always returns a `Token` object with a special `EOF` type, and `Lexer.advance` will do nothing.

`Token` objects store their type, their string value, and their location in the source code.

```
@dataclass
class Token:
  type: TokenType
  value: str
  location: Location
```

The lexer works by keeping an index of the current position in the source code string and a copy of the current token. `Lexer.token` just returns the current token; the real work is done in `Lexer.advance`. `Lexer.advance` examines the current character of the string to determine what to do. For instance, if it is a double quote, the method will read to the end of a string literal. If it is a digit, it will read an integer token. If the character is ambiguous (for example, `<` could be a single operator or the start of the `<=` compound operator), the lexer may need to "peek" ahead one or more characters.

The lexer contains the logic of handling backslash escapes in string literals. For example, `"\n"` is two characters in the source code (excluding the quotes) but should produce a token that is only one character in length (the ASCII character `0xA`).

### The parser
The parser transforms the stream of tokens emitted by the lexer into a hierarchical parse tree.

```
class Parser:
  def parse(self, lexer: Lexer) -> Node:
    ...
```

It is written as a [recursive descent parser](https://en.wikipedia.org/wiki/Recursive_descent_parser). For each type of parse tree node, the parser has a corresponding `match` method that consumes an instance of that node from the lexer. For example, here is a sample implementation of the `match_while_loop` method:

```
def match_while_loop(self) -> Node:
  assert lexer.token().type == TokenType.WHILE
  lexer.advance()

  condition = self.match_expression()
  body = self.match_block()

  return WhileNode(condition, body)
```

The invariant of all `match` methods is that when the method is called, `lexer.token()` will return the first token of the node, and when the method returns, it will leave `lexer.token()` at the first token of the next node. This makes it easy to string together `match` methods as above, where `match_block` is called directly after `match_expression`.

Infix expressions (including function calls, for which the left parenthesis can be viewed as the infix operator) are parsed with the Pratt parsing technique and a precedence table. An example of this techniqu can be seen in [this GitHub gist](https://gist.github.com/iafisher/5ef9aef5e04e376d59ff80f743b0a38e).

### Parse trees
The parse tree directly represents the structure of the program as it was written by the programmer. Each type of syntactic structure is represented by a separate Python class. Here is the definition of the `WhileNode` class:

```
@dataclass
class WhileNode(Statement):
  condition: Expression
  body: List[Statement]
```

All classes inherit from one of three abstract base classes: `Expression`, `Statement`, and `Type`.

## Semantic analysis
### The analyzer
TODO

### Abstract syntax trees and types
TODO

Primitive types are represented by constant objects: `VENICE_TYPE_I64`, `VENICE_TYPE_STRING`, etc. Compound types are represented by Python classes, e.g.:

```
@dataclass
class ListType(Type):
  element_type: Type
```

## Code generation
### ASTs to CFGs
TODO

```
@dataclass
class Function:
  name: str
  parameters: List[Parameter]
  return_type: VILType
  basic_blocks: OrderedDict[str, Block]

@dataclass
class Block:
  instructions: List[Instruction]
  exit: ExitInstruction
```

### CFGs to machine code
TODO

#### Complications
- VIL has infinite number of registers while all real machines have a fixed, low number.
- Some operations only operate on certain registers, e.g. `idiv` only works on `rdx` and `rax`.
- Variadic functions require setting `al` to the number of vector registers used.

## The Venice Intermediate Language
See <https://github.com/iafisher/venice/blob/master/docs/vil.md> for the specification of the Venice Intermediate Language.

## Venice runtime
TODO

## Sample compilations
### Hello, world
Source program:

```
func main() -> i64 {
  println("Hello, world!");
  return 0;
}
```

Parse tree:

```
(func
  main
  ()
  (type i64)
  (
    (call println ((string "Hello, world!")))
    (return (int 0))
  )
)
```

Abstract syntax tree:

```
(func
  main
  ()
  (type i64)
  (
    (call:void println ((string:string "Hello, world!")))
    (return (int:i64 0))
  )
)
```

VIL code:

```
func venice_println(a0 ptr<u8>) -> void;

const s0:ptr<u8> = "Hello, world!"

func main() -> i64 {
  call venice_println s0:ptr<u8>
  ret 0:i64
}
```

TODO: Should the type of `s0` be a pointer or a fixed-size array?

TODO: At what point does the `println` global get resolved to a `venice_println` external runtime function?

x86 code:

```
extern venice_println

section .text
global main

main:
  mov rdi, s0
  call venice_println

  mov rax, 0
  ret

section .data
  s0 db "Hello, world!"
```

### `if` statements
Source program:

```
let x: i64 = 0;
if (do_something() > 10) {
  x = 10;
} else {
  x = 11;
}
```

Parse tree:

```
(
  (let x (type i64) (int 0))
  (if
    (> (call do_something ()) (int 10))
    ((assign x (int 10)))
    ((assign x (int 11)))
  )
)
```

Abstract syntax tree:

```
(
  (let x (type i64) (int:i64 0))
  (if
    (>:bool (call do_something ()) (int:i64 10))
    ((assign x (int:i64 10)))
    ((assign x (int:i64 11)))
  )
)
```

VIL code:

```
l0:
  t0 = alloca:i64 8
  store t0, 0
  t1 = call do_something
  t2 = cmp_gt t1, 10
  br t2, l1, l2

l1:
  store t0, 10
  br l3

l2:
  store t0, 11
  br l3

l3:
  ...
```

x86 code:

```
  mov rbp, rsp
  sub rsp, 8
  mov [rbp - 8], 0
  call do_something
  cmp rax, 10
  bg l1
  br l2

l1:
  mov [rbp - 8], 10
  br l3

l2:
  mov [rbp - 8], 11
  br l3

l3:
  ...
```

## Further reading
- <https://www.llvm.org/docs/tutorial/MyFirstLanguageFrontend/index.html>
- <https://mapping-high-level-constructs-to-llvm-ir.readthedocs.io/en/latest/README.html>
