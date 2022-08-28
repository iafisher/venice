# The Venice implementation
WORK IN PROGRESS

## Introduction
This document describes the design of the compiler for the Venice programming language. For a description of the language itself, see the [language reference](https://github.com/iafisher/venice/blob/master/docs/language-reference.md).

## Goals
The focus of the first version of the compiler, which will be written in Rust, is correctness and simplicity. Efficiency, either in the sense of emitting efficient code or compiling programs quickly, is not a focus. Error messages for incorrect programs should give enough information for the programmer to be able to identify the error (i.e., the location in the source code and a clear description of the problem), but they do not need to be as detailed as, e.g., Rust's.

These goals may change in future versions of the compiler.

## Compiler architecture
The fundamental task of the compiler is to translate a Venice program written in textual source code into an executable file in x86-64 assembly code.

For simplicity, the first version of the compiler will emit a textual assembly-language program which an external assembler and linker can turn into an actual binary executable. Future versions of the compiler may directly produce ELF binaries.

Translation occurs in three phases:

- Lexical and syntactic analysis
- Semantic analysis
- Code generation

...and involves four different representations of the source program:

- Parse tree
- Abstract syntax tree
- Linear intermediate representation
- x86 machine code

The output of the lexical and syntactic analysis phase is a concrete parse tree whose nodes correspond directly to textual elements in the program's source code. For example, if there is a `while` loop in the source code, there will be a corresponding `while` node in the parse tree.

The parse tree is converted into an abstract syntax tree, which has a similar structure but simplifies some syntactic forms, and adds type information to each node. For example, `if` statements in Venice can have an arbitrary number of `else if` clauses, but in the abstract syntax tree there are only `if-else` nodes and `else if` clauses are represented by nesting two or more `if-else` nodes.

The abstract syntax tree is lowered into a linear intermediate representation using an abstract assembly language called VIL. See docs/vil.md for a description of VIL.

In the last step, the machine-independent linear representation is converted into concrete machine code for a specific architecture (in Venice's case, x86-64).
