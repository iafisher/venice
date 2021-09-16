package bytecode

import (
	"fmt"
)

type CompiledProgram struct {
	Version int
	Imports []*CompiledProgramImport
	Code    map[string][]Bytecode
}

type CompiledProgramImport struct {
	Path string
	Name string
}

func NewCompiledProgram() *CompiledProgram {
	return &CompiledProgram{
		Version: 1,
		Code: map[string][]Bytecode{
			"main": []Bytecode{},
			// TODO(2021-08-26): Think of a better way to do this.
			"Optional__Some": []Bytecode{
				&PushEnum{"Some", 1},
			},
		},
	}
}

type Bytecode interface {
	fmt.Stringer
	bytecode()
}

// Pop x, y from the stack and push x + y.
type BinaryAdd struct{}

// Pop x, y from the stack and push x ++ y where x and y are strings.
type BinaryConcat struct{}

// Pop x, y from the stack and push x == y.
type BinaryEq struct{}

// Pop x, y from the stack and push x > y.
type BinaryGt struct{}

// Pop x, y from the stack and push x >= y.
type BinaryGtEq struct{}

// Pop x, y from the stack and push x in y.
type BinaryIn struct{}

// Pop x, y from the stack and push x[y] where x is a list.
type BinaryListIndex struct{}

// Pop x, y from the stack and push x < y.
type BinaryLt struct{}

// Pop x, y from the stack and push x <= y.
type BinaryLtEq struct{}

// Pop x, y from the stack and push x[y] where x is a map.
type BinaryMapIndex struct{}

// Pop x, y from the stack and push x % y.
type BinaryModulo struct{}

// Pop x, y from the stack and push x * y.
type BinaryMul struct{}

// Pop x, y from the stack and push x != y.
type BinaryNotEq struct{}

// Same as BinaryAdd, BinaryMul, etc. except the operands are real numbers instead of
// integers.
type BinaryRealAdd struct{}
type BinaryRealDiv struct{}
type BinaryRealMul struct{}
type BinaryRealSub struct{}

// Pop x, y from the stack and push x[y] where x is a string.
type BinaryStringIndex struct{}

// Pop x, y from the stack and push x - y.
type BinarySub struct{}

// Pop N values from the stack, bundle them into a class object, and push it onto the stack.
type BuildClass struct {
	Name string
	N    int
}

// Pop N values from the stack, bundle them into a list, and push it onto the stack.
type BuildList struct {
	N int
}

// Pop N*2 values from the stack where each key is before its corresponding value
// (e.g., [K, V, K, V]), bundle them into a map, and push it onto the stack.
type BuildMap struct {
	N int
}

// Pop N values from the stack, bundle them into a tuple, and push it onto the stack.
type BuildTuple struct {
	N int
}

// Pop a function and N arguments from the top of stack (e.g., [arg3, arg2, arg1, f]),
// call the function with the arguments, and push the return value onto the stack if it
// is not void.
type CallFunction struct {
	N int
}

// Check the top of the stack, which must be an enum object, and push true if it has the
// given label, or false otherwise.
//
// Used to implement pattern-matching.
type CheckLabel struct {
	Name string
}

// Duplicate the value at the top of the stack. This only copies the reference, not the
// value itself.
type DupTop struct{}

// Call Next() on the top of the stack, which must be an iterator. If it returns nil, jump
// N instructions forward in the program. Otherwise, push the return value onto the stack.
type ForIter struct {
	N int
}

// Pop the top of the stack, convert it into an iterator, and push the iterator onto the
// stack.
type GetIter struct{}

// Push a function object representing the method called `Name` on the value on the top of
// the stack.
type LookupMethod struct {
	Name string
}

// Placeholder instruction used internally by the compiler.
type Placeholder struct {
	Name string
}

// Push a constant boolean value onto the stack.
type PushConstBool struct {
	Value bool
}

// Push a constant function object onto the stack.
type PushConstFunction struct {
	Name      string
	IsBuiltin bool
}

// Push a constant integer value onto the stack.
type PushConstInt struct {
	Value int
}

// Push a constant real number value onto the stack.
type PushConstRealNumber struct {
	Value float64
}

// Push a constant string value onto the stack.
type PushConstStr struct {
	Value string
}

// Push a constant enum object onto the stack.
type PushEnum struct {
	Name string
	N    int
}

// Assuming the top of the stack is an enum object, push the value of the enum at `Index`.
//
// Note that unlike PushField and PushTupleField, this instruction leaves the enum on the
// stack.
type PushEnumIndex struct {
	Index int
}

// Assuming the top of the stack is a class object, pop it and push its field at `Index`.
//
// Note that unlike PushEnumIndex, this instruction pops the class object from the stack.
type PushField struct {
	Index int
}

// Push the value of the given symbol onto the stack.
type PushName struct {
	Name string
}

// Assuming the top of the stack is a tuple, pop it and push its field at `Index`.
//
// Note that unlike PushEnumIndex, this instruction pops the tuple from the stack.
type PushTupleField struct {
	Index int
}

// Jump N instructions forward in the program.
//
// If N = 0, it loops forever. If N = 1, it is a no-op. If N > 1, it skips over 1 or more
// instructions.
type RelJump struct {
	N int
}

// Pop the top of the stack, which must be a boolean value. If it is false, jump N
// instructions forward in the program. Otherwise, continue as normal.
type RelJumpIfFalse struct {
	N int
}

// Pop the top of the stack, which must be a boolean value. If it is false, jump N
// instructions forward in the program. Otherwise, pop the top of the stack and
// continue as normal.
type RelJumpIfFalseOrPop struct {
	N int
}

// Pop the top of the stack, which must be a boolean value. If it is true, jump N
// instructions forward in the program. Otherwise, pop the top of the stack and
// continue as normal.
type RelJumpIfTrueOrPop struct {
	N int
}

// Return from the current function.
type Return struct{}

// Pop x, y, z from the stack and push z, x, y.
type RotThree struct{}

// Pop x, y from the stack and store y in x[Index].
type StoreField struct {
	Index int
}

// Pop x, y, z from the stack and store z in x[y] where x is a list.
type StoreIndex struct{}

// Pop x, y, z from the stack and store z in x[y] where x is a map.
type StoreMapIndex struct{}

// Pop a value from the stack and store it under `Name` in the current environment.
type StoreName struct {
	Name string
}

// Pop x from the stack and push -x.
type UnaryMinus struct{}

// Pop x from the stack and push not x.
type UnaryNot struct{}

// Assuming the top of the stack is a tuple, pop it and push its contents individually
// onto the stack, e.g. if the stack is [..., x] and x = (a, b, c), then the stack will be
// [..., c, b, a].
type UnpackTuple struct{}

func (b *BinaryAdd) String() string {
	return "BINARY_ADD"
}

func (b *BinaryConcat) String() string {
	return "BINARY_CONCAT"
}

func (b *BinaryEq) String() string {
	return "BINARY_EQ"
}

func (b *BinaryGt) String() string {
	return "BINARY_GT"
}

func (b *BinaryGtEq) String() string {
	return "BINARY_GT_EQ"
}

func (b *BinaryIn) String() string {
	return "BINARY_IN"
}

func (b *BinaryListIndex) String() string {
	return "BINARY_LIST_INDEX"
}

func (b *BinaryLt) String() string {
	return "BINARY_LT"
}

func (b *BinaryLtEq) String() string {
	return "BINARY_LT_EQ"
}

func (b *BinaryMapIndex) String() string {
	return "BINARY_MAP_INDEX"
}

func (b *BinaryModulo) String() string {
	return "BINARY_MODULO"
}

func (b *BinaryMul) String() string {
	return "BINARY_MUL"
}

func (b *BinaryNotEq) String() string {
	return "BINARY_NOT_EQ"
}

func (b *BinaryRealAdd) String() string {
	return "BINARY_REAL_ADD"
}

func (b *BinaryRealDiv) String() string {
	return "BINARY_REAL_DIV"
}

func (b *BinaryRealMul) String() string {
	return "BINARY_REAL_MUL"
}

func (b *BinaryRealSub) String() string {
	return "BINARY_REAL_SUB"
}

func (b *BinaryStringIndex) String() string {
	return "BINARY_STRING_INDEX"
}

func (b *BinarySub) String() string {
	return "BINARY_SUB"
}

func (b *BuildClass) String() string {
	return fmt.Sprintf("BUILD_CLASS %q %d", b.Name, b.N)
}

func (b *BuildList) String() string {
	return fmt.Sprintf("BUILD_LIST %d", b.N)
}

func (b *BuildMap) String() string {
	return fmt.Sprintf("BUILD_MAP %d", b.N)
}

func (b *BuildTuple) String() string {
	return fmt.Sprintf("BUILD_TUPLE %d", b.N)
}

func (b *CallFunction) String() string {
	return fmt.Sprintf("CALL_FUNCTION %d", b.N)
}

func (b *CheckLabel) String() string {
	return fmt.Sprintf("CHECK_LABEL %q", b.Name)
}

func (b *DupTop) String() string {
	return "DUP_TOP"
}

func (b *ForIter) String() string {
	return fmt.Sprintf("FOR_ITER %d", b.N)
}

func (b *GetIter) String() string {
	return "GET_ITER"
}

func (b *LookupMethod) String() string {
	return fmt.Sprintf("LOOKUP_METHOD %q", b.Name)
}

func (b *Placeholder) String() string {
	return fmt.Sprintf("PLACEHOLDER %q", b.Name)
}

func (b *PushConstBool) String() string {
	if b.Value {
		return "PUSH_CONST_BOOL 1"
	} else {
		return "PUSH_CONST_BOOL 0"
	}
}

func (b *PushConstFunction) String() string {
	var x string
	if b.IsBuiltin {
		x = "1"
	} else {
		x = "0"
	}

	return fmt.Sprintf("PUSH_CONST_FUNCTION %q %s", b.Name, x)
}

func (b *PushConstInt) String() string {
	return fmt.Sprintf("PUSH_CONST_INT %d", b.Value)
}

func (b *PushConstRealNumber) String() string {
	return fmt.Sprintf("PUSH_CONST_REAL_NUMBER %f", b.Value)
}

func (b *PushConstStr) String() string {
	return fmt.Sprintf("PUSH_CONST_STR %q", b.Value)
}

func (b *PushEnum) String() string {
	return fmt.Sprintf("PUSH_ENUM %q %d", b.Name, b.N)
}

func (b *PushEnumIndex) String() string {
	return fmt.Sprintf("PUSH_ENUM_INDEX %d", b.Index)
}

func (b *PushField) String() string {
	return fmt.Sprintf("PUSH_FIELD %d", b.Index)
}

func (b *PushName) String() string {
	return fmt.Sprintf("PUSH_NAME %q", b.Name)
}

func (b *PushTupleField) String() string {
	return fmt.Sprintf("PUSH_TUPLE_FIELD %d", b.Index)
}

func (b *RelJump) String() string {
	return fmt.Sprintf("REL_JUMP %d", b.N)
}

func (b *RelJumpIfFalse) String() string {
	return fmt.Sprintf("REL_JUMP_IF_FALSE %d", b.N)
}

func (b *RelJumpIfFalseOrPop) String() string {
	return fmt.Sprintf("REL_JUMP_IF_FALSE_OR_POP %d", b.N)
}

func (b *RelJumpIfTrueOrPop) String() string {
	return fmt.Sprintf("REL_JUMP_IF_TRUE_OR_POP %d", b.N)
}

func (b *Return) String() string {
	return "RETURN"
}

func (b *RotThree) String() string {
	return "ROT_THREE"
}

func (b *StoreField) String() string {
	return fmt.Sprintf("STORE_FIELD %d", b.Index)
}

func (b *StoreIndex) String() string {
	return "STORE_INDEX"
}

func (b *StoreMapIndex) String() string {
	return "STORE_MAP_INDEX"
}

func (b *StoreName) String() string {
	return fmt.Sprintf("STORE_NAME %q", b.Name)
}

func (b *UnaryMinus) String() string {
	return "UNARY_MINUS"
}

func (b *UnaryNot) String() string {
	return "UNARY_NOT"
}

func (b *UnpackTuple) String() string {
	return "UNPACK_TUPLE"
}

func (b *BinaryAdd) bytecode()           {}
func (b *BinaryConcat) bytecode()        {}
func (b *BinaryEq) bytecode()            {}
func (b *BinaryGt) bytecode()            {}
func (b *BinaryGtEq) bytecode()          {}
func (b *BinaryIn) bytecode()            {}
func (b *BinaryListIndex) bytecode()     {}
func (b *BinaryLt) bytecode()            {}
func (b *BinaryLtEq) bytecode()          {}
func (b *BinaryMapIndex) bytecode()      {}
func (b *BinaryModulo) bytecode()        {}
func (b *BinaryMul) bytecode()           {}
func (b *BinaryNotEq) bytecode()         {}
func (b *BinaryRealAdd) bytecode()       {}
func (b *BinaryRealDiv) bytecode()       {}
func (b *BinaryRealMul) bytecode()       {}
func (b *BinaryRealSub) bytecode()       {}
func (b *BinaryStringIndex) bytecode()   {}
func (b *BinarySub) bytecode()           {}
func (b *BuildClass) bytecode()          {}
func (b *BuildList) bytecode()           {}
func (b *BuildMap) bytecode()            {}
func (b *BuildTuple) bytecode()          {}
func (b *CallFunction) bytecode()        {}
func (b *CheckLabel) bytecode()          {}
func (b *DupTop) bytecode()              {}
func (b *ForIter) bytecode()             {}
func (b *GetIter) bytecode()             {}
func (b *LookupMethod) bytecode()        {}
func (b *Placeholder) bytecode()         {}
func (b *PushConstBool) bytecode()       {}
func (b *PushConstInt) bytecode()        {}
func (b *PushConstFunction) bytecode()   {}
func (b *PushConstRealNumber) bytecode() {}
func (b *PushConstStr) bytecode()        {}
func (b *PushEnum) bytecode()            {}
func (b *PushEnumIndex) bytecode()       {}
func (b *PushField) bytecode()           {}
func (b *PushName) bytecode()            {}
func (b *PushTupleField) bytecode()      {}
func (b *RelJump) bytecode()             {}
func (b *RelJumpIfFalse) bytecode()      {}
func (b *RelJumpIfFalseOrPop) bytecode() {}
func (b *RelJumpIfTrueOrPop) bytecode()  {}
func (b *Return) bytecode()              {}
func (b *RotThree) bytecode()            {}
func (b *StoreField) bytecode()          {}
func (b *StoreIndex) bytecode()          {}
func (b *StoreMapIndex) bytecode()       {}
func (b *StoreName) bytecode()           {}
func (b *UnaryMinus) bytecode()          {}
func (b *UnaryNot) bytecode()            {}
func (b *UnpackTuple) bytecode()         {}
