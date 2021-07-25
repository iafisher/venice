# The Venice programming language
- Author: Ian Fisher (iafisher@fastmail.com)
- Status: draft
- Last updated: July 25, 2021


**NOTE**: Nothing described in this document has been implemented yet.


## Atomic values
### Integers, real numbers and booleans
Integers can be written in decimal (`42`), hexadecimal (`0x2a`), octal (`0o52`), and binary (`0b101010`). Leading zeroes are not allowed for unprefixed integer literals, so that C-style octal numbers are not misinterpreted as decimal numbers.

Real numbers can be written in standard decimal notation (`3.14`) or in scientific notation (`1e10`).

The `integer` and `real` types in Venice are arbitrary precision. Fixed-width signed and unsigned integer types are also available: `i64`, `i32`, `i16`, `i8`, `u64`, `u32`, `u16`, and `u8`. Standard IEEE 754 single and double precision floating-point number types are available as `float` and `double`.

<!-- TODO(2021-07-22): Implicit and explicit coercion of integer and real types. -->

The two boolean literals are written as `true` and `false`.

### Strings and characters
String literals are enclosed in double quotes (`"hello, world"`). Standard backslash escapes are supported:

- `\'` for a single quote (although single quotes do not need to be backslash-escaped)
- `\"` for a double quote
- `\\` for a backslash
- `\n` for a newline
- `\r` for a carriage return
- `\t` for a tab
- `\b` for a backspace
- `\f` for a form feed
- `\v` for a vertical tab
- `\0` for a null character
- `\xFF` for the Unicode character with hex value `FF`
- `\uxxxx` for the Unicode character with 16-bit hex value `xxxx`
- `\Uxxxxxxxx` for the Unicode character with 32-bit hex value `xxxxxxxx`

String literals may contain backslashes. The following string literal is equal to `"hello\n world"`:

```venice
"hello
 world"
```

If the newline is immediately preceded by a backslash, then leading whitespace is stripped from the next line. The following string literal is equal to `"hello world"`:

```venice
"hello \
         world"
```

Expressions can be interpolated into Venice strings using `${...}`, e.g. `1 + 1 = ${1 + 1}`.

Character literals are enclosed in single quotes (`'a'`) and support the same set of backslash escapes as string literals.

### Lists and tuples
A list stores an ordered sequence of elements of the same type (`[1, 2, 3]`).

A tuple stores a fixed number of elements, potentially of different types (`(1, "two", 3.0)`).

### Maps and sets
A map associates keys with values, with efficient look-up, insertion and deletion (`{1: "one", 2: "two", 3: "three"}`). All keys must be of the same type, and all values must be of the same type, but the key type and value type may differ.

A set stores an unordered collection of elements of the same type with no duplicates (`{1, 2, 3}`).


## Declarations
### Variables
The `let` statement is used to declare a symbol with an initial value.

```venice
let x: int = 10
```

The symbol's type can be declared explicitly as above, but it can also be omitted and inferred by the compiler.

```venice
let y = "abc"  // inferred type of "string"
```

Bindings with `let` are permanent. If you need to reassign to a symbol later, use `var` for the initial declaration instead.

```venice
var x = 10
x = 9
```

### Functions
Function declarations use the `fn` keyword. Unlike `let` and `var` declarations, function declarations require types to be listed.

```venice
fn add(x: integer, y: integer) -> integer {
  return x + y
}

fn print_plus_two(x: integer) -> void {
  print(x)
}
```

Functions can have named parameters and default parameter values. All parameters after the special `*` symbol must be specified by name when calling the function.

```venice
fn pluralize(word: string, *, ending: string = "s") -> string {
  return word + ending;
}

print(pluralize("match", ending = "es"))  // matches
```


## Control flow
### `if` statements
Venice has standard `if`-`else if`-`else` statements.

```venice
if minutes < 10 {
  print("short wait")
} else if minutes < 20 {
  print("medium wait")
} else {
  print("long wait")
}
```

### `while` loops
The `while` loop is used to loop while a condition is true.

```venice
let i = 0;
while i < 10 {
  print(i)
  i += 1
}
```

### `for` loops
The `for` loop is used to iterate over the elements of a sequence.

```venice
let letters: list<string> = ["a", "b", "c"];
for letter in letters {
  print(letter)
}
```


## Advanced data types
### Classes
Classes in Venice are similar to structs in C and Rust (and unlike classes in object-oriented languages like Python and Java).

```venice
class User {
  public name: string
  public age: integer
}

let u = User(name = "John Doe", age = 24)
print(u.name)  // John Doe
u.age += 1
print(u.age)  // 25
```

A constructor is generated for classes by default, and objects can be compared for equality, hashed, and printed as long as all their constituent types can be.

Methods can be defined on a class.

```venice
class User {
  public name: string
  public age: integer

  public as_string(self) -> string {
    return "${self.name}, aged ${self.age}"
  }
}
```

Methods and fields must be declared as either `public` or `private`.

### Algebraic data types
Venice supports algebraic data types (ADTs).

```venice
enum Expression {
  InfixOperation(op: string, left: Expression, right: Expression),
  Integer(integer),
  String(string),
}

let e = Expression::InfixOperation(
  op = "+",
  left = Expression::Integer(20),
  right = Expression::Integer(22),
}
```

Class-like types declared inside ADTs can also be used as independent types as if they were regular classes.

```venice
fn print_infix(e: Expression::InfixOperation) {
  print("${e.left} ${e.op} ${e.right}")
}
print_infix(e)
```

ADTs can also be used as simple C-style enums.

```venice
enum Direction { Up, Down, Left, Right }
```

### Pattern matching
The `match` statement is used to pattern-match ADTs. Patterns are tested in order from top to bottom. The listed patterns must be exhaustive. If necessary, a catch-all `default` pattern can be added to the bottom.

```venice
match e {
  // The 'Expression::' prefix is unnecessary inside a match statement.
  case InfixOperation(op, ... ) {
    print("Infix operation: ${op}")
  },
  default {
    print("Some other thing.")
  },
}

enum InputEvent {
  MouseClick(x: integer, y: integer),
  Key(code: integer, shift: boolean, ctrl: boolean),
  Fn(integer),
  Esc,
}

match x {
  // Match a class-like subtype.
  case MouseClick(x, y) {
  },
  // Match part of a type's fields.
  case Key(code, ...) {
  },
  // Match regular subtypes.
  case Fn(x) {
  },
  case Esc {
  },
}
```

Like in Rust, if the last line of each clause of a match statement is an expression of equivalent types, the match statement overall can be used as an expression.

Pattern matching can also be done in `if let` statements like in Rust.

```venice
if let Key(code, ...) = event {
  print(code)
}
```

### Interfaces
Interfaces are used to encapsulate related objects with the same interface but different implementations.

```venice
interface StringLike {
  as_string() -> string
}
```

Interfaces must be implemented explicitly using the `for` keyword in the method definition.

```venice
class Foo {
  public x: integer

  public as_string(self) -> string for StringLike {
    return "Foo(${self.x})"
  }
}
```

Methods implementing an interface must be `public`.

Interface types can be used for function parameters (and anywhere else that types are used).

```venice
fn print_anything(x: StringLike) {
  print(x.as_string())
}
```

### Generics
Structs and ADTs can be made generic over one or more types.

```venice
enum Optional<T> {
  Some(T),
  None
}
```

Generics may be constrained. In the example below, whatever type substitutes for `C` must implement the interface `StringLike`.

```venice
class Collection<A, B, C: StringLike> {
  public a: A
  public b: B
  public c: C
}
```


## Formal syntax
```bnf
program := import* declaration+
import  := IMPORT SYMBOL

declaration := function_declaration
             | enum_declaration
             | class_declaration
             | variable_declaration

function_declaration := FN SYMBOL LPAREN (parameter_list | parameter_list_with_asterisk)? RPAREN (ARROW type)? block

parameter_list := (parameter COMMA)* parameter COMMA?
parameter      := SYMBOL COLON type (EQ expression)?

parameter_list_with_asterisk := (parameter COMMA)* ASTERISK COMMA (parameter COMMA)* parameter COMMA?

enum_declaration := ENUM SYMBOL type_parameter_list? LCURLY (enum_case COMMA)* enum_case COMMA? RCURLY
enum_case        := SYMBOL (LPAREN (parameter_list | symbol_list) RPAREN)?
symbol_list      := (SYMBOL COMMA)* SYMBOL COMMA?

class_declaration := CLASS SYMBOL type_parameter_list? LPAREN parameter_list RPAREN class_body
class_body        := LCURLY function_declaration* RCURLY
class_method      := SYMBOL LPAREN SELF (COMMA (parameter_list | parameter_list_with_asterisk))? RPAREN (ARROW type)? block

variable_declaration := (LET | CONST) SYMBOL (COLON type)? EQ expression

block     := LCURLY NEWLINE (statement NEWLINE)* RCURLY
statement := while | for | if | return | assign | match | BREAK | CONTINUE | expression

while  := WHILE expression block
for    := FOR symbol IN expression block
if     := IF condition BLOCK elif* else?
elif   := ELSE IF condition block
else   := ELSE block
return := RETURN expression
assign := SYMBOL (EQ | PLUS_EQ | MINUS_EQ | TIMES_EQ | DIV_EQ) expression
match  := MATCH expression LCURLY (match_case COMMA)+ (DEFAULT block COMMA?)? RCURLY

condition     := expression | LET match_pattern EQ expression
match_case    := CASE match_pattern BLOCK
match_pattern := SYMBOL | symbol_list (COMMA ELLIPSIS)?

expression := atom | prefix | infix | call | LPAREN expression RPAREN
prefix     := OP expression
infix      := expression OP expression
call       := expression LPAREN (argument_list | argument_list_with_keywords)? RPAREN

argument_list := (expression COMMA)* expression COMMA?
argument_list_with_keywords := (argument_list COMMA)? (SYMBOL EQ expression COMMA)* SYMBOL EQ expression COMMA?

type := SYMBOL type_parameter_list?
type_paramater_list := LANGLE symbol_list RANGLE
```


## Implementation
The reference implementation of Venice comprises two programs: `vnc`, which compiles Venice programs to bytecode, and `vnvm`, which runs bytecode programs.

### The Venice bytecode format
TODO(2021-07-04)
