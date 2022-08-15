# The Venice Intermediate Language
**DRAFT**

The Venice Intermediate Language (VIL) is an [intermediate representation](https://en.wikipedia.org/wiki/Intermediate_representation) used internally by the Venice compiler. It is similar to [LLVM's IR](https://llvm.org/docs/LangRef.html).

A VIL program consists of one or more function declarations. Each function declaration in turn consists of one or more basic blocks. A basic block is a sequence of straight-line (i.e., non-branching) instructions, terminated by a single branching instruction.

## The VIL virtual machine
VIL programs executed on an abstract virtual machine that is an idealization of real processors.

The VIL VM has an unlimited number of numbered registers as well as main memory. Each register and individual memory location holds a 64-bit signed integer in two's complement. Memory locations are byte-addressable; words are little-endian.

The only formal difference between registers and memory is that registers are not addressed. The optimizer assumes that registers are faster than memory.

The VM has a single binary comparison flag whose value is set by the `cmp` family of instructions.

## Instructions
VIL instructions have the general form:

```
<destination> = <instruction> <operand1> <operand2>
```

For instance,

```
%3 = add %0 %1
```

Most instructions operate solely on registers. A small set of instructions are used to interact with main memory.

### Memory access
The `alloca` instruction is used to declare named memory locations.

```
%result = alloca:i64 8
```

Its operand is the number of bytes on the stack that the value will occupy; it must be a constant. `%result` is declared to be a pointer to type `i64`. 

The `load` instruction allows you to load the contents of a memory location into a register. It takes a pointer argument and an index (which may be a constant or a register).

```
%0 = load %result, 0
```

The `store` instruction allows you to save the contents of a register into a memory location. Its types are flipped compared to `load`: the destination is a pointer and the first operand is a register.

```
%result = store %0, 1
```

### Arithmetic and bitwise operations
TODO

### Comparison and branching
- `cmpeq/cmpneq/cmplt/cmpgt/cmplte/cmpgte <r1> <r2>`: Sets the comparison flag to 1 if the instruction's condition holds between the two operands, to 0 otherwise.
- `jump <label>`: Jump unconditionally to the label.
- `jumpif <label1>, <label2>`: Jumps to the first label if the comparison flag is set to 1, to the second label otherwise.
