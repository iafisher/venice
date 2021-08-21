package bytecode

import (
	"fmt"
)

type CompiledProgram struct {
	Version int
	Code    map[string][]Bytecode
}

func NewCompiledProgram() *CompiledProgram {
	return &CompiledProgram{
		Version: 1,
		Code:    map[string][]Bytecode{"main": []Bytecode{}},
	}
}

type Bytecode interface {
	fmt.Stringer
	bytecode()
}

type BinaryAdd struct{}

type BinaryAnd struct{}

type BinaryConcat struct{}

type BinaryDiv struct{}

type BinaryEq struct{}

type BinaryGt struct{}

type BinaryGtEq struct{}

type BinaryIn struct{}

type BinaryListIndex struct{}

type BinaryLt struct{}

type BinaryLtEq struct{}

type BinaryMapIndex struct{}

type BinaryMul struct{}

type BinaryNotEq struct{}

type BinaryOr struct{}

type BinaryStringIndex struct{}

type BinarySub struct{}

// TODO(2021-08-15): Combine this and ContinueLoop into a Placeholder instruction.
type BreakLoop struct{}

type BuildClass struct {
	Name string
	N    int
}

type BuildList struct {
	N int
}

type BuildMap struct {
	N int
}

type BuildTuple struct {
	N int
}

type CallBuiltin struct {
	N int
}

type CallFunction struct {
	N int
}

type ContinueLoop struct{}

type ForIter struct {
	N int
}

type GetIter struct{}

type LookupMethod struct {
	Name string
}

type PushConstBool struct {
	Value bool
}

type PushConstChar struct {
	Value byte
}

type PushConstInt struct {
	Value int
}

type PushConstStr struct {
	Value string
}

type PushEnum struct {
	Name string
	N    int
}

type PushField struct {
	Index int
}

type PushName struct {
	Name string
}

type PushTupleField struct {
	Index int
}

type RelJump struct {
	N int
}

type RelJumpIfFalse struct {
	N int
}

type RelJumpIfFalseOrPop struct {
	N int
}

type RelJumpIfTrueOrPop struct {
	N int
}

type Return struct{}

type StoreField struct {
	Index int
}

type StoreIndex struct{}

type StoreName struct {
	Name string
}

type UnaryMinus struct{}

type UnaryNot struct{}

func (b *BinaryAdd) String() string {
	return "BINARY_ADD"
}

func (b *BinaryAnd) String() string {
	return "BINARY_AND"
}

func (b *BinaryConcat) String() string {
	return "BINARY_CONCAT"
}

func (b *BinaryDiv) String() string {
	return "BINARY_DIV"
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

func (b *BinaryMul) String() string {
	return "BINARY_MUL"
}

func (b *BinaryNotEq) String() string {
	return "BINARY_NOT_EQ"
}

func (b *BinaryOr) String() string {
	return "BINARY_OR"
}

func (b *BinaryStringIndex) String() string {
	return "BINARY_STRING_INDEX"
}

func (b *BinarySub) String() string {
	return "BINARY_SUB"
}

func (b *BreakLoop) String() string {
	return "BREAK_LOOP"
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

func (b *CallBuiltin) String() string {
	return fmt.Sprintf("CALL_BUILTIN %d", b.N)
}

func (b *CallFunction) String() string {
	return fmt.Sprintf("CALL_FUNCTION %d", b.N)
}

func (b *ContinueLoop) String() string {
	return "CONTINUE_LOOP"
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

func (b *PushConstBool) String() string {
	if b.Value {
		return "PUSH_CONST_BOOL 1"
	} else {
		return "PUSH_CONST_BOOL 0"
	}
}

func (b *PushConstChar) String() string {
	return fmt.Sprintf("PUSH_CONST_CHAR %q", string(b.Value))
}

func (b *PushConstInt) String() string {
	return fmt.Sprintf("PUSH_CONST_INT %d", b.Value)
}

func (b *PushConstStr) String() string {
	return fmt.Sprintf("PUSH_CONST_STR %q", b.Value)
}

func (b *PushEnum) String() string {
	return fmt.Sprintf("PUSH_ENUM %q %d", b.Name, b.N)
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

func (b *StoreField) String() string {
	return fmt.Sprintf("STORE_FIELD %d", b.Index)
}

func (b *StoreIndex) String() string {
	return "STORE_INDEX"
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

func (b *BinaryAdd) bytecode()           {}
func (b *BinaryAnd) bytecode()           {}
func (b *BinaryConcat) bytecode()        {}
func (b *BinaryDiv) bytecode()           {}
func (b *BinaryEq) bytecode()            {}
func (b *BinaryGt) bytecode()            {}
func (b *BinaryGtEq) bytecode()          {}
func (b *BinaryIn) bytecode()            {}
func (b *BinaryListIndex) bytecode()     {}
func (b *BinaryLt) bytecode()            {}
func (b *BinaryLtEq) bytecode()          {}
func (b *BinaryMapIndex) bytecode()      {}
func (b *BinaryMul) bytecode()           {}
func (b *BinaryNotEq) bytecode()         {}
func (b *BinaryOr) bytecode()            {}
func (b *BinaryStringIndex) bytecode()   {}
func (b *BinarySub) bytecode()           {}
func (b *BreakLoop) bytecode()           {}
func (b *BuildClass) bytecode()          {}
func (b *BuildList) bytecode()           {}
func (b *BuildMap) bytecode()            {}
func (b *BuildTuple) bytecode()          {}
func (b *CallBuiltin) bytecode()         {}
func (b *CallFunction) bytecode()        {}
func (b *ContinueLoop) bytecode()        {}
func (b *ForIter) bytecode()             {}
func (b *GetIter) bytecode()             {}
func (b *LookupMethod) bytecode()        {}
func (b *PushConstBool) bytecode()       {}
func (b *PushConstChar) bytecode()       {}
func (b *PushConstInt) bytecode()        {}
func (b *PushConstStr) bytecode()        {}
func (b *PushEnum) bytecode()            {}
func (b *PushField) bytecode()           {}
func (b *PushName) bytecode()            {}
func (b *PushTupleField) bytecode()      {}
func (b *RelJump) bytecode()             {}
func (b *RelJumpIfFalse) bytecode()      {}
func (b *RelJumpIfFalseOrPop) bytecode() {}
func (b *RelJumpIfTrueOrPop) bytecode()  {}
func (b *Return) bytecode()              {}
func (b *StoreField) bytecode()          {}
func (b *StoreIndex) bytecode()          {}
func (b *StoreName) bytecode()           {}
func (b *UnaryMinus) bytecode()          {}
func (b *UnaryNot) bytecode()            {}
