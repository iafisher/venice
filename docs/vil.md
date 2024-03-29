# The Venice Intermediate Language
**DRAFT**

The Venice Intermediate Language (VIL) is an [intermediate representation](https://en.wikipedia.org/wiki/Intermediate_representation) used internally by the Venice compiler. It is similar to [LLVM's IR](https://llvm.org/docs/LangRef.html).

A VIL program consists of one or more function declarations. Each function declaration in turn consists of one or more basic blocks. A basic block is a sequence of straight-line (i.e., non-branching) instructions, terminated by a single branching instruction.

## The VIL virtual machine
VIL programs run on an abstract virtual machine that is an idealization of real processors. In practice the VIL machine is similar to an x86 processor.

The VIL machine has 16 registers:

- 6 registers for passing arguments to functions, named `%rp0` through `%rp5`
- 7 general-purpose registers, named `%rg0` through `%rg6`
- 3 special registers: `%rt` to store a function's return value, `%rsp` to store the current stack pointer, and `%rbp` to store the current base pointer

Each register and individual memory location holds a 64-bit signed integer in two's complement. Memory locations are byte-addressable; words are little-endian. Registers are not addressable.

## Instructions
VIL instructions have the general form:

```
<destination> = <instruction> <operand1> <operand2>
```

For instance,

```
%rg3 = add %rg0, %rg1
```

Most instructions operate solely on registers (which are represented in VIL code as a number preceded by a percent sign). A small set of instructions are used to interact with main memory.

In general, VIL instruction fall into one of several classes:

- Instructions that map more or less directly to x86, such as `add` and `sub`.
- Instructions that map to several x86 instructions that form a conceptual unit, such as `frame_set_up` and `frame_tear_down`.
- Instructions that express the intent of the corresponding x86 instruction, such as `callee_save` and `callee_restore`.

### Register manipulation
The `set` instruction places a constant value in a register:

```
%rg0 = set 42
```

The `move` instruction copies a value from one register to another:

```
%rg1 = move %rg0
```

### Memory access
The `load` instruction allows you to load the contents of a memory location into a register. It takes an offset from the top of the stack frame.

```
%rg0 = load -8
```

The `store` instruction allows you to save the contents of a register into a memory location. Like `store`, it takes an offset from the top of the stack frame.

```
store %rg0, -8
```

### Arithmetic and bitwise operations
TODO

### Comparison and branching
- `cmp <r1> <r2>`: Compares `r1` and `r2`, setting flags but discarding the result.
- `jump_XYZ <label1> <label2>`: Jumps to `label1` if the condition indicated by `XYZ` holds, or to `label2` otherwise.
- `jump <label>`: Jumps unconditionally to the label.

### Functions
The `call` instruction is used to call a function. It jumps to the function's body and arranges the stack so that a subsequent `ret` instruction will return to the place the function was called from. It expects that its parameters have been placed into the parameter registers.

```
%rp0 = move %rg0
%rp1 = move %rg1
call add
%rg2 = move %rt
```

All parameter registers are implicitly caller-save (meaning the function is free to use them and the caller must save them and restore them after the call if it wishes to preserve their value), as are `%rt`, `%rsp`, `%rg0`, and `%rg1`. The other general-purpose registers and `%rbp` are callee-save.

### Constants
TODO

### Static data
TODO
