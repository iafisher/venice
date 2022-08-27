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
The `alloca` instruction is used to declare named memory locations.

```
%result = alloca:i64 8
```

Its operand is the number of bytes on the stack that the value will occupy; it must be a constant. `%result` is declared to be a pointer to type `i64`. 

The `load` instruction allows you to load the contents of a memory location into a register. It takes a pointer argument and an index (which may be a constant or a register).

```
%rg0 = load %result, 0
```

The `store` instruction allows you to save the contents of a register into a memory location. Its types are flipped compared to `load`: the destination is a pointer and the first operand is a register.

```
%result = store %rg0, 1
```

### Arithmetic and bitwise operations
TODO

### Comparison and branching
- `cmpeq/cmpneq/cmplt/cmpgt/cmplte/cmpgte <r1> <r2>`: Sets the comparison flag to 1 if the instruction's condition holds between the two operands, to 0 otherwise.
- `jump <label>`: Jump unconditionally to the label.
- `jumpif <label1>, <label2>`: Jumps to the first label if the comparison flag is set to 1, to the second label otherwise.

### Functions
The `call` instruction is used to call a function. It jumps to the function's body and arranges the stack so that a subsequent `ret` instruction will return to the place the function was called from. It expects that its parameters have been placed into the parameter registers.

```
%rp0 = move %rg0
%rp1 = move %rg1
%rg2 = call add
```

All parameter registers are implicitly caller-save (meaning the function is free to use them and the caller must save them and restore them after the call if it wishes to preserve their value), as are `%rt`, `%rsp`, `%rg0`, and `%rg1`. The other general-purpose registers and `%rbp` are callee-save.

### Constants
TODO

### Static data
TODO
