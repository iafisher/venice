# Venice 0.1 implementation
Ian Fisher, July 2022

WORK IN PROGRESS

## Introduction
This document describes the implementation of version 0.1 of the Venice language. For a description of the language itself, see the [version 0.1 design](https://github.com/iafisher/venice/blob/master/docs/design/venice-v0.1.md).

## Compiler architecture
The fundamental task of the compiler is to translate a Venice program written in textual source code into an executable file in x86-64 assembly code.

For simplicity, the first version of the compiler will emit a textual assembly-language program which an external assembler can turn into an actual binary executable, but future versions of the compiler will directly emit ELF binaries.

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
