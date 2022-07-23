# Venice 0.1
Ian Fisher, July 2022


## Introduction
This document describes version 0.1 of the Venice programming language. It is not intended as a tutorial.

Version 0.1 is designed to be minimal but still complete enough that useful medium-sized programs can be written in it. While I expect Venice to change significantly as it approaches version 1.0, the core of the language should remain recognizable, if not wholly identical. An abridged list of expected changes to the language can be found in the "Future additions" section.


## Overview of the Venice language
Venice is a modern, high-level, statically-typed programming language that compiles to x86-64 machine code. Among popular existing languages, Venice is most similar to Python, although it is statically typed and has a C-style syntax. It is at a similar level of abstraction as Java, and higher than Rust, C, and C++. For instance, a Venice programmer does not have access to raw pointers, and memory management is automatic. When ease of use comes into conflict with efficiency, ease of use is usually preferred. For more background on the motivation behind Venice, see my blog post ["Why am I writing a new programming language"](https://iafisher.com/blog/2021/09/why-i-am-writing-a-new-programming-language).


## Expressions
### Booleans and integers
Booleans and integers in Venice are similar to other languages. The only numeric type is `i64`, which represents a 64-bit signed integer.[^may-change] The names of the boolean literals are spelled `true` and `false` and are case-sensitive. There is no implicit conversion between booleans and integers.

Floating-point numbers, unsigned integer types, and integer types of different bit widths are not supported.[^may-change]

### Strings
A string is an immutable sequence of Unicode characters. In the text of the program, string literals are delimited by double quotes:

```
let s: string = "abc";
```

The characters between the double quotes must be valid UTF-8 text; this is implicitly enforced by the overall requirement that Venice programs are encoded in UTF-8. Several backslash escapes are allowed:

- `\\` for a literal backslash
- `\"` for a double quote that does not terminate the string
- `\n` for a newline
- `\r` for a carriage return
- `\t` for a horizontal tab
- `\xhh` for the ASCII character with hex value `hh` (both digits must be included)
- `\uxxxx` for the Unicode character with 16-bit hex value `xxxx` (all four digits must be included)
- `\Uxxxxxxxx` for the Unicode character with 32-bit hex value `xxxxxxxx` (all eight digits must be included)

String literals may not contain raw newlines or carriage returns.

Strings may be indexed:

```
assert s[0] == "a";
```

Like in Python, the elements of a string are themselves one-character strings. Venice has no separate character type.

If the index is out of bounds, the program is terminated with an error message.[^may-change] This is because version 0.1 lacks facilities for error-handling.

The `size` method returns the number of Unicode characters in the string.

```
assert s.size() == 3;
```

The `codepoint` method returns the Unicode code point, as an integer, of the first character of the string. If the string is empty, `0` is returned.[^may-change]

### Lists
A list is an ordered sequence of values of the same type. List literals use the same syntax as Python:

```
let l: list<i64> = [1, 2, 3];
```

Like strings, lists can be indexed and have their `size` taken:

```
assert l[0] == 1;
assert l.size() == 3;
```

If the index is out of bounds, the program is terminated with an error message.[^may-change]

Elements can be appended to the end of the list, inserted at an arbitrary point, or removed:

```
l.append(4);
l.insert(0, -1);  # the first argument is the index, the second the element
assert l.remove(1) == 1;
```

List elements can also be directly assigned:

```
l[0] = 1;
```

### Tuples
A tuple is an ordered sequence type, like a list, with two key differences: the length of a tuple is fixed at compile time, and the elements of the tuple can be of different types. For instance:

```
let employee: tuple<i64, string> = (123, "John Doe");
```

Elements of a tuple are accessed with a different syntax:

```
assert employee.0 == 123;
assert employee.1 == "John Doe";
```

The index must be an integer constant; it cannot be an arbitrary integer expression or variable, because its value must be deducible at compile time for type-checking.

Tuples are immutable, so the individual elements of a tuple cannot be re-assigned after creation. Since the size of a tuple is fixed, there is no `size()` method on tuples.

### Maps
A map is a mapping from keys to values. All keys in a map must be of the same type, and likewise for values, but the key type and value type may be different.

```
let numbers: map<i64, string> = {1: "one", 2: "two", 3: "three"};
```

Only primitive types (booleans, integers, and strings) can be the keys of a map,[^may-change] but any type can be a value.

Values are set and retrieved from maps with the same syntax as lists:

```
m[4] = "four";
assert m[4] == "four";
```

If the value is not found in the map, the program is terminated with an error message.[^may-change]

Maps have a number of useful methods:

```
assert m.contains(3);
m.remove(3);
assert not m.contains(3);
assert m.size() == 3;
```

### Records
A record is a collection of a fixed number of named and typed fields. It is the only user-defined data type in Venice 0.1. Record types are declared with the `record` keyword:

```
record User {
  name: string,
  visit_count: i64,
  is_admin: bool,
};
```

Record types must be declared at the top-level of a program, i.e. not inside a function.

The `new` keyword is used to create an instance of a record type:

```
let user: User = new User { name: "John Doe", visit_count: 0, is_admin: false };
```

All fields must be explicitly specified by name.[^may-change] The order does not need the order of the fields in the record type's declaration.

Fields can be accessed and re-assigned using the dot operator:

```
user.name = "Jane Doe";
assert user.name == "Jane Doe";
```

### Boolean operations
The standard boolean operators—`and`, `or`, and `not`—are supported in Venice 0.1. Boolean operators are lazy, e.g. the expression `x and y` won't result in the evaluation of `y` if `x` evaluates to false.

### Comparisons
The operators `>`, `>=`, `<`, and `<=` can be used to compare integers and strings. `==` and its inverse `!=` can be used to check equality of any two values of the same type. For compound types (lists, tuples, maps, and records), the equality operator first checks that the containers have the same number of elements (if applicable—two tuples or two records of the same type will always have the same number of elements), then recursively checks that each element is equal to the other. All comparison operators yield a boolean value.

### Arithmetic operations
The arithmetic operators in Venice are `+`, `-`, `*`, `//`, and `%`, for addition, subtraction, multiplication, floor division, and modulo, respectively. The `/` operator is reserved for future use as the real division operator with floating-point numbers. `-` also serves as the unary negation operator. The result of an arithmetic operation that causes integer overflow or underflow is undefined.[^may-change]

Venice 0.1 has no bitwise operators, and the internal representation of integers (e.g., two's complement) is left unspecified.[^may-change]

### Concatenation operator
`++`, the concatenation operator, is used to concatenate two strings or lists together:

```
assert "abc" ++ "def" == "abcdef";
```

If the operands are lists, they must have the same element type. The resulting list is a shallow copy of the original lists. This means that changes to the list itself (adding and removing elements) will not affect the original lists, but internal changes to the elements will.

### Function calls
Functions are called with the normal C-style syntax:

```
let result: i64 = add(1, 2);
```

All arguments are evaluated before being passed to the function. The order of evaluation is not specified, so passing expressions whose evaluation has side effects should be avoided.

All values are passed by reference in Venice; implicit copies of function arguments are never made. For mutable types (lists, maps, and records), any changes made in the function will be visible to the caller.

### Built-in functions
- `i64(s: string) -> i64`: Converts the string to an integer. If the string cannot be converted, then the minimum value of a 64-bit signed integer (`-2^63`) is returned.[^may-change]
- `print(s: string) -> void`: Prints the string to standard output, followed by a newline.
- `readline(prompt: string) -> string`: Prints the prompt to standard output (without adding a newline) and then reads returns a line from standard input.
- `string(x: any) -> string`: Converts a value of any type to a string representation.

### Operator precedence
The precedence of infix and prefix operators in Venice is as follows, from lowest precedence (least binding) to highest precedence (most binding):

- `or`
- `and`
- `not`
- Comparison operators
- `+`, `++`, and `-`
- `*`, `//`, and `%`
- `-` unary operator
- Indexing, function calls, and attribute access
- Parenthesized expressions


## Control flow
### Variable declaration
The `let` keyword is used to declare a variable with an initial value.

```
let x: i64 = 0;
```

The type declaration is mandatory.[^may-change] It is a compile error to declare a variable when a variable of the same name is already in scope. The scope of a variable is the nearest block (delimited by curly braces) that encloses its `let` declaration. In most cases, this is the variable's enclosing function, although it could be an `if` clause, a `while` loop, or any other kind of block. The scope of a function's parameters is the entire function. The scope of a function itself is the entire program file; functions and variables exist in the same namespace.

### Variable assignment
A variable declared with a `let` statement can be assigned to a new value with the `=` operator:

```
x = 1;
```

The new value must be of the same type as the variable's declared type.

### If-else statements
Conditional control flow is implemented with `if-else` statements:

```
if x == 0 {
  handle_zero(x);
} else if x > 0 {
  handle_positive_number(x);
} else {
  handle_negative_number(x);
}
```

The only mandatory part of an `if-else` statement is the `if` clause itself, which may be followed by zero or more `else if` clauses and a single, optional `else` clause.

Like in Python and unlike in C, parentheses around the `if` condition are optional.


### While loops
The simplest kind of loop is the `while` loop, which repeats its body until its condition is no longer true:

```
let counter: i64 = 0;
while counter < 10 {
  counter += 1;
}
```

A `while` loop may contain `break` and `continue` statements to exit the loop early or continue immediately to the next iteration, respectively.


### For loops
`for` loops are used to iterate over sequence types: strings, lists, and maps.

```
let names: list<string> = ["John", "Jill", "Jane"];
for name in names {
  print(name);
}
```

The scope of the variable introduced in the `for` loop is the entire body of the loop.

Looping over a map requires two variables, one for the key and one for the value:

```
let numbers: map<i64, string> = {1: "one", 2: "two"};
for key, value in numbers {
  print(string(key) ++ " is " ++ value);
}
```

It is undefined behavior to modify a sequence while looping over it.


## Declarations
### Constants
Constant values may be declared at the top-level of a program with the `const` keyword:

```
const MAX_FILES: i64 = 256;
```

An upper-case name is recommended as a stylistic matter but is not mandated by the language. Constant variables cannot be re-assigned, and their initial value must be a compile-time constant, i.e. either a primitive value (boolean, integer, or string) or a list, tuple, or map containing only compile-time constants.

### Functions
A function is declared with the `func` keyword. The types of a function's parameters and its return type must be explicitly annotated.

```
func add(x: i64, y: i64) -> i64 {
  return x + y;
}
```

All paths out of a function must end with a `return` statement. For instance, if the last statement of a function is an `if-else` statement, then both branches must have a `return` statement at the end. The exception is functions with a return type of `void`; while an explicit `return` statement can be used to exit early, the `return` statement at the end is optional.


## Programs
A Venice program is a single file of UTF-8 text. The top-level of a Venice program consists of one or more declarations of constants, types, and functions. There must be at least one function declared with the name `main` which takes either no parameters or a single parameter of type `list<string>`, and has a return type of either `i64` or `void`.


## Type-checking
Every expression in a Venice program has a concrete type that can be deduced at compile time. Various contexts require expressions of a particular type; for example, the `+` requires that its operands be integers. These conditions are checked at compile time.

Types `t1` and `t2` are the same if:

- They are both the same primitive type (`bool`, `i64`, or `string`).
- They are both list types and their element types are the same.
- They are both tuple types and each of their element types are the same.
- They are both map types and their key types are the same and their value types are the same.
- They are both record types of the same name.

Note that the final condition means that two record types declared separately but with identical fields are considered distinct types.


## Memory management
Memory is automatically garbage-collected in Venice; that is, memory is returned to the system once no live variable in the running program can access it. The exact mechanism is left unspecified.


## Future additions
This is a list of the most important additions I plan to make to Venice. It is neither comprehensive nor binding, but it should give the reader a general idea of the direction in which the language is heading.

- Multiple-file programs
- An expanded standard library
- A mechanism to signal and handle errors
- Limited type inference (for example, in `let` statements)
- Algebraic data types and pattern matching
- String interpolation
- Floating-point numbers
- First-class and anonymous functions


[^may-change]: This may change in a future version of Venice.
