# An Introduction to the Venice Programming Language
Venice is a modern, high-level, statically-typed programming language. It is designed to be powerful, intuitive, and easy to learn. This tutorial introduces the reader to the fundamental concepts and features of Venice. It assumes that the reader has working knowledge of at least one other programming language.

**NOTE:** Venice is currently in the early stages of development, and not all of the features described in this tutorial have been implemented.


## A note on notation
This tutorial is illustrated with a large number of real Venice code snippets, formatted as they would be typed into Venice's interactive read-eval-print loop (REPL). Lines beginning with `>>>` represent the programmer's input, and other lines represent the REPL's output.

```
>>> 1 + 1
2
```

Some expressions and statements span multiple lines. These are indicated by ellipses:

```
>>> print(
...   1 + 1
... )
...
2
```

The REPL supports a few commands beginning with exclamation points, such as `!type`, which will be introduced later on in this tutorial. Keep in mind that these are not part of the Venice language and will not work outside of the REPL.


## Running Venice code
Once you have followed the installation instructions to get your own copy of Venice, you can run Venice code in two ways. First, you can invoke the `venice` command with no arguments to start a REPL where you can enter Venice code one line at a time:

```shell
$ venice
The Venice programming language.

>>> 1 + 1
2
```

The REPL is great for quick calculations and experimentation, but it is less suitable for developing real programs. Instead, you can write Venice code in your favorite text editor, and use `venice execute` to run it:

```shell
$ venice execute my_program.vn
...
```

`venice execute` compiles and executes the program in one step. If you wish to compile but not execute it, use `venice compile` instead. This command will produce a `.vnb` file, which can then be passed to `venice execute` later to execute it.


## Values
Now that you are up and running with Venice, let's explore some of Venice's core features. As you might expect, Venice has the usual set of basic types, including integers, real numbers, booleans, characters, and strings:

```
>>> 1 + 1
2
>>> 3.14159
3.14159
>>> true or false
true
>>> 'A'
'A'
>>> "Venice" ++ " is great"
"Venice is great"
```

Lists represent an ordered sequence of objects of the same type:

```
>>> [1, 2, 3]
[1, 2, 3]
```

Maps represent a mapping from objects of one type to another:

```
>>> {"Spain": "Madrid", "France": "Paris", "Germany": "Berlin"}
{"Spain": "Madrid", "France": "Paris", "Germany": "Berlin"}
```

Tuples represent a fixed sequence of objects of possibly different types:

```
>>> (1, "two", 3.0)
(1, "two", 3.0)
```

Sets represent an unordered sequence of unique objects of the same type:

```
>>> {1, 2, 2, 3, 3, 3}
{1, 2, 3}
```

If you are coming from a Python background, then lists, maps, tuples, and sets will work exactly as you expect. Note that, like in Python, lists are named after the common English term and not the technical term "linked list". They function like arrays or vectors in other languages.


## Static typing
Venice is statically typed. This means that it is possible to deduce the types of all values and symbols in the program without running the program. Many languages do not have this property. For example, in the below Python program, the type of the symbol `x` depends on what `random.random()` returns, which cannot be known without running it.

```python
import random

if random.random() < 0.5:
  x = 0
else:
  x = "a string"
```

Static typing has the major advantage that it can catch programming errors without running the code. Its major disadvantage is that it requires more work on the part of the programmer.

Some statically-typed languages, like Java and C++, are notorious for requiring redundant and verbose type annotations. Others, notably Haskell, use advanced techniques to type-check the program while allowing the programmer to omit most or all type annotations. Venice seeks a middle ground between the two: type annotations can often be omitted, but must be supplied in certain critical places, such as at function definitions.

Using the `!type` command in the REPL, we can discover the type of any expression:

```
>>> !type 1
int
>>> !type [1, 2, 3]
[int]
>>> !type ("one", 2)
(string, int)
```


## Control flow
The `let` keyword is used to bind values to symbols:

```
>>> let name = "Venice"
>>> name
"Venice"
```

Symbols declared with `let` cannot subsequently be re-assigned to. If you need to re-assign to a symbol, use `var` instead:

```
>>> var i = 0
>>> i += 1
>>> i
1
```

If you wish, you can explicitly declare the type of the symbol with `let` or `var`, although the compiler can usually infer it for you:

```
>>> let name2: string = "Venice"
```

Venice supports standard imperative control flow structures, including `if` statements:

```
>>> let role = "user"
>>> if role == "admin" {
...   print("Logged in as admin.")
... } else if role == "user" {
...   print("Logged in.")
... } else {
...   print("Not logged in.")
... }
...
Logged in.
```

...`while` loops:

```
>>> var i = 0
>>> while i < 10 {
...   print(i)
...   i += 2
... }
...
0
2
4
6
8
```

...and `for` loops:

```
>>> let islands = ["Madagascar", "Borneo", "Honshu", "Cuba"]
>>> for island in islands {
...   print(island)
... }
...
Madagascar
Borneo
Honshu
Cuba
```

Unlike their equivalents in some languages like C and Java, `for` loops in Venice are exclusively used to iterate over sequences of values, like lists, maps, and strings. To count up or down in a `for` loop, use the `range` function:

```
>>> for i in range(5) {
...   print(i)
... }
...
0
1
2
3
4
```

Indentation is not significant in Venice programs, however statements must end with a newline character.


## Functions
Functions are defined with the `fn` keyword:

```
>>> fn fibonacci(n: uint) -> uint {
...   if n == 0 {
...     return 0
...   } else if n == 1 or n == 2 {
...     return 1
...   } else {
...     return fibonacci(n - 1) + fibonacci(n - 2)
...   }
... }
...
>>> fibonacci(12)
144
```

The types of a function's parameters and its return type must be explicitly annotated. This both simplifies the implementation of the type-checker, and provides useful documentation for the users of your functions.

Functions may require keyword arguments with the pseudo-paramater `*`, and specify default values:

```
>>> fn pluralize(n: int, word: string, *, plural_form: string = "") -> string {
...   let s = (
...     word
...     if n == 1
...     else (
...       plural_form
...       if plural_form != ""
...       else
...       word ++ "s"
...     )
...   )
...   return "${n} ${s}"
... }
...
>>> pluralize(2, "child", plural_form = "children")
"2 children"
```

The implementation of `pluralize` demonstrates a few other Venice features. `if-else` can be used as an expression, equivalent to the ternary conditional operator in C-like languages. The syntax `${...}` interpolates values into strings.


## Types
TODO: introduction

- Built-in types like `int` and `list`
- Custom types defined with `class`
- Algebraic data types defined with `enum`
- Interface types defined with `interface`



## Classes
The `class` keyword is used to define custom data types:

```
>>> class Point {
...  public x: int
...  public y: int
...
...  public fn swap(self) {
...    let tmp = self.x
...    self.x = self.y
...    self.y = tmp
... }
...
```

Classes come with a constructor, string representation, and equality operation for free:

```
>>> let p = Point(x = 1, y = 2)
>>> print(p)
Point(x = 1, y = 2)
>>> p.x
1
>>> p.swap()
>>> print(p)
Point(x = 2, y = 1)
>>> p == Point(x = 1, y = 1)
false
```

TODO: constructors


## Algebraic data types and pattern matching
TODO: Introduction to algebraic data types

The `enum` keyword is used to define algebraic data types. In its most basic usage, it can define a simple set of labels, like in C:

```
>>> enum Direction { Up, Down, Left, Right }
>>> Direction::Up
Direction::Up
```

Unlike in C, each label can be associated with data:

```
>>> enum Json {
...   JsonObject({string: Json}),
...   JsonArray([Json]),
...   JsonString(string),
...   JsonNumber(int),
...   JsonBool(bool),
...   JsonNull,
... }
...
```

This usage of `enum` is closer to union types in C, but unlike union types, `enum` types are type-safe in Venice.

Inner values of `enum` types can be accessed with pattern matching through the `match` statement:

```
>>> match j {
...   case JsonArray([x, ...]) {
...     print(x)
...   }
...   default {}
... }
...
1
```

Matches must be exhaustive, which is why the empty `default` clause above is required.

Pattern matching is also possible with the `if let` statement:

```
>>> let j = Json::JsonArray([Json::JsonNumber(1), Json::JsonNumber(2)])
>>> if let Json::JsonArray(array) = j {
...   print(length(array))
... }
...
3
```


## Interfaces
Interface types represent types with the same fixed set of methods but with different implementations:

```
>>> interface HasSize {
...   fn size() -> uint
... }
...
```

Interfaces are satisfied by classes implicitly. In the example below, `Rectangle` satisfies `HasSize` by virtue of its `size` method.

```
>>> class Rectangle {
...   public width: int
...   public height: int
...
...   public fn size(self) -> uint {
...     return self.width * self.height
...   }
... }
...
```

Interfaces may be used wherever regular types can be:

```
>>> fn add_sizes(o1: HasSize, o2: HasSize) -> uint {
...   return o1.size() + o2.size()
... }
...
```


## Generics
Functions, classes, and `enum` types may take one or more generic type parameters.

```
>>> fn check_equals<T>(x: T, y: T) {
...   if x != y {
...     print("${x} != ${y}")
...   }
... }
...
```

TODO: generics constrained by interfaces


## Wrapping up
TODO
