# The Venice programming language
- Author: Ian Fisher (iafisher@fastmail.com)
- Status: draft
- Last updated: July 4, 2021


**NOTE**: Nothing described in this document has been implemented yet.


Venice is a high-level, statically-typed programming language.

<!-- TODO(2021-07-04): Put a concise but demonstrative example of Venice code here. -->


## A tour of Venice
```venice
/**
 * Numbers
 *
 * Integers and real numbers are arbitrary precision, unless the floating-point
 * format is explicitly requested with the 'f' suffix.
 */
42
0x2a
0o52
0b101010
1e7
// A decimal number
3.14159
// A floating-point number
3.14159f


/**
 * Booleans
 */
true
false


/**
 * Strings
 */

"a single line string"
// Strings may contain newlines.
"hello
  world"
// Long strings can be broken over multiple lines with a backslash.
// The two strings below are equivalent. (Note that leading whitespace is
// stripped from the second line, but trailing whitespace is preserved on the
// first line.)
"a \\
 b"
"a b"
// Single quotes are used for characters.
'a'
// Venice supports string interpolation.
"1 + 1 = ${1 + 1}"


/**
 * Lists
 *
 * All elements of a list must be of the same type.
 */
[1, 2, 3]


/**
 * Tuples
 *
 * A tuple may contain elements of different types, but its length is fixed.
 */
(1, "two", 3.0)


/**
 * Maps
 */
{1: "one", 2: "two", 3: "three"}


/**
 * Declarations and assignments
 */
let x: int = 10
// The symbol's type can be declared explicitly, or inferred by the compiler.
let y = "abc"
// Let bindings are permanent. If you need to reassign to a symbol, use 'var'
// instead.
var x = 10
x = 9


/**
 * Functions
 *
 * A function declaration must list the types of its parameters and its return
 * value.
 */
fn add(x: integer, y: integer) -> integer {
  return x + y
}

// A function without an explicit return type has an implicit return type of
// "void".
fn print_plus_two(x: integer) {
  print(x)
}

print_plus_two(40)  // 42

// Functions can have named parameters and default parameter values.
fn pluralize(word: string, *, ending: string = "s") -> string {
  return word + ending;
}
print(pluralize("match", ending = "es"))  // matches


/**
 * Control flow
 */
let minutes = 30;
if minutes < 10 {
  print("short wait")
} else if minutes < 20 {
  print("medium wait")
} else {
  print("long wait")
}

let i = 0;
while i < 10 {
  print(i)
  i += 1
}

let letters: list<string> = ["a", "b", "c"];
for letter in letters {
  print(letter)
}


/**
 * Structs
 */
struct User(name: string, age: integer) {
  as_string(self) -> string {
    return "${self.name}, aged {self.age}"
  }
}

let u = User(name = "John Doe", age = 24)
print(u.name)  // John Doe
u.age += 1
print(u.age)  // 25


/**
 * Algebraic data types
 */
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

match e {
  // The 'Expression::' prefix is unnecessary inside a match statement.
  case InfixOperation(op, ... ) {
    print("Infix operation: ${op}")
  },
  default {
    print("Some other thing.")
  },
}

// Structs declared inside ADTs can also be used as independent types.
fn print_infix(e: Expression::InfixOperation) {
  print("${e.left} ${e.op} ${e.right}")
}
print_infix(e)


/**
 * Pattern matching
 */
enum InputEvent(
  MouseClick(x: integer, y: integer),
  Key(code: integer, shift: boolean, ctrl: boolean),
  Fn(integer),
  Esc,
)

match x {
  // Match a struct subtype.
  case MouseClick(x, y) {
  },
  // Match part of a struct's fields.
  case Key(code, ...) {
  },
  // Match regular subtypes.
  Fn(x) {
  },
  Esc {
  },
  // Match statements must be exhaustive.
}
// Like in Rust, if the last line of each clause of a match statement is an
// expression of equivalent types, the match statement overall can be used as
// an expression.

// Pattern matching can also be done in 'if let' statements like in Rust.
if let Key(code, ...) = event {
  print(code)
}

/**
 * Interfaces
 */
interface StringLike {
  as_string() -> string,
}

struct Foo(x: integer) {
  as_string(self) -> string {
    return "Foo(${self.x})"
  }
}

fn print_anything(x: StringLike) {
  print(x.as_string())
}

/**
 * Generics
 */
enum Optional<T>(
  Some(T),
  None
)

// Type C must implement the interface StringLike.
struct Collection<A, B, C: StringLike>(a: A, b: B, c: C)
```


## Language reference
### Syntax
```bnf
program := import* declaration+
import  := IMPORT SYMBOL

declaration := function_declaration
             | enum_declaration
             | struct_declaration
             | variable_declaration

function_declaration := FN SYMBOL LPAREN (parameter_list | parameter_list_with_asterisk)? RPAREN (ARROW type)? block

parameter_list := (parameter COMMA)* parameter COMMA?
parameter      := SYMBOL COLON type (EQ expression)?

parameter_list_with_asterisk := (parameter COMMA)* ASTERISK COMMA (parameter COMMA)* parameter COMMA?

enum_declaration := ENUM SYMBOL type_parameter_list? LCURLY (enum_case COMMA)* enum_case COMMA? RCURLY
enum_case        := SYMBOL (LPAREN (parameter_list | symbol_list) RPAREN)?
symbol_list      := (SYMBOL COMMA)* SYMBOL COMMA?

struct_declaration := STRUCT SYMBOL type_parameter_list? LPAREN parameter_list RPAREN struct_body
struct_body        := LCURLY function_declaration* RCURLY
struct_method      := SYMBOL LPAREN SELF (COMMA (parameter_list | parameter_list_with_asterisk))? RPAREN (ARROW type)? block

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
The Venice implementation comprises two programs: `vnc`, which compiles Venice programs to bytecode, and `vvm`, which runs bytecode programs.

### The Venice bytecode format
TODO(2021-07-04)
