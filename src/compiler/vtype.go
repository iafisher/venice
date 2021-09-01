package compiler

import (
	"fmt"
	"strings"
)

type VeniceType interface {
	fmt.Stringer
	veniceType()
	GetGenericParameters() []string
}

/**
 * Compound types
 */

type VeniceAnyType struct{}

type VeniceBooleanType struct{}

type VeniceCharacterType struct{}

type VeniceClassType struct {
	Name              string
	GenericParameters []string
	Fields            []*VeniceClassField
	Methods           []*VeniceFunctionType
}

// Helper struct - does not implement VeniceType
type VeniceClassField struct {
	Name      string
	Public    bool
	FieldType VeniceType
}

type VeniceEnumType struct {
	Name              string
	GenericParameters []string
	Cases             []*VeniceCaseType
}

// Helper struct - does not implement VeniceType
type VeniceCaseType struct {
	Label string
	Types []VeniceType
}

type VeniceFunctionType struct {
	Name              string
	Public            bool
	GenericParameters []string
	ParamTypes        []VeniceType
	ReturnType        VeniceType
	IsBuiltin         bool
}

type VeniceIntegerType struct{}

type VeniceListType struct {
	ItemType VeniceType
}

type VeniceMapType struct {
	KeyType   VeniceType
	ValueType VeniceType
}

type VeniceModuleType struct {
	Name  string
	Types map[string]VeniceType
}

type VeniceRealNumberType struct{}

type VeniceStringType struct{}

type VeniceSymbolType struct {
	Label string
}

type VeniceTupleType struct {
	ItemTypes []VeniceType
}

type VeniceUnionType struct {
	Types []VeniceType
}

/**
 * Primitive type declarations
 */

var (
	VENICE_TYPE_ANY         = &VeniceAnyType{}
	VENICE_TYPE_BOOLEAN     = &VeniceBooleanType{}
	VENICE_TYPE_CHARACTER   = &VeniceCharacterType{}
	VENICE_TYPE_INTEGER     = &VeniceIntegerType{}
	VENICE_TYPE_REAL_NUMBER = &VeniceRealNumberType{}
	VENICE_TYPE_STRING      = &VeniceStringType{}
	VENICE_TYPE_OPTIONAL    = &VeniceEnumType{
		Name:              "Optional",
		GenericParameters: []string{"T"},
		Cases: []*VeniceCaseType{
			&VeniceCaseType{
				"Some",
				[]VeniceType{&VeniceSymbolType{"T"}},
			},
			&VeniceCaseType{"None", nil},
		},
	}
)

func VeniceOptionalTypeOf(concreteType VeniceType) VeniceType {
	return &VeniceEnumType{
		Name:              "Optional",
		GenericParameters: []string{},
		Cases: []*VeniceCaseType{
			&VeniceCaseType{
				"Some",
				[]VeniceType{concreteType},
			},
			&VeniceCaseType{"None", nil},
		},
	}
}

/**
 * String() implementations
 */

func (t *VeniceAnyType) String() string {
	return "any"
}

func (t *VeniceBooleanType) String() string {
	return "boolean"
}

func (t *VeniceCharacterType) String() string {
	return "char"
}

func (t *VeniceClassType) String() string {
	return t.Name
}

func (t *VeniceEnumType) String() string {
	return t.Name
}

func (t *VeniceFunctionType) String() string {
	var sb strings.Builder
	sb.WriteString("func(")
	for i, param := range t.ParamTypes {
		sb.WriteString(param.String())
		if i != len(t.ParamTypes)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte(')')

	if t.ReturnType != nil {
		sb.WriteString(" -> ")
		sb.WriteString(t.ReturnType.String())
	}

	return sb.String()
}

func (t *VeniceIntegerType) String() string {
	return "int"
}

func (t *VeniceListType) String() string {
	return fmt.Sprintf("[%s]", t.ItemType.String())
}

func (t *VeniceMapType) String() string {
	return fmt.Sprintf("{%s: %s}", t.KeyType.String(), t.ValueType.String())
}

func (t *VeniceModuleType) String() string {
	return t.Name
}

func (t *VeniceRealNumberType) String() string {
	return "real"
}

func (t *VeniceStringType) String() string {
	return "string"
}

func (t *VeniceSymbolType) String() string {
	return t.Label
}

func (t *VeniceTupleType) String() string {
	var sb strings.Builder
	sb.WriteByte('(')
	for i, itemType := range t.ItemTypes {
		sb.WriteString(itemType.String())
		if i != len(t.ItemTypes)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte(')')
	return sb.String()
}

func (t *VeniceUnionType) String() string {
	var sb strings.Builder
	sb.WriteString("union<")
	for i, subType := range t.Types {
		sb.WriteString(subType.String())
		if i != len(t.Types)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte('>')
	return sb.String()
}

/**
 * GetGenericParameters implementations
 */

func (t *VeniceAnyType) GetGenericParameters() []string {
	return nil
}

func (t *VeniceBooleanType) GetGenericParameters() []string {
	return nil
}

func (t *VeniceCharacterType) GetGenericParameters() []string {
	return nil
}

func (t *VeniceClassType) GetGenericParameters() []string {
	return nil
}

func (t *VeniceEnumType) GetGenericParameters() []string {
	return t.GenericParameters
}

func (t *VeniceFunctionType) GetGenericParameters() []string {
	return nil
}

func (t *VeniceIntegerType) GetGenericParameters() []string {
	return nil
}

func (t *VeniceListType) GetGenericParameters() []string {
	return nil
}

func (t *VeniceMapType) GetGenericParameters() []string {
	return nil
}

func (t *VeniceModuleType) GetGenericParameters() []string {
	return nil
}

func (t *VeniceRealNumberType) GetGenericParameters() []string {
	return nil
}

func (t *VeniceStringType) GetGenericParameters() []string {
	return nil
}

func (t *VeniceSymbolType) GetGenericParameters() []string {
	return nil
}

func (t *VeniceTupleType) GetGenericParameters() []string {
	return nil
}

func (t *VeniceUnionType) GetGenericParameters() []string {
	return nil
}

/**
 * Miscellaneous methods
 */

func (t *VeniceCaseType) AsFunctionType(enumType *VeniceEnumType) *VeniceFunctionType {
	return &VeniceFunctionType{
		Name:              fmt.Sprintf("%s::%s", enumType.Name, t.Label),
		Public:            true,
		GenericParameters: enumType.GenericParameters,
		ParamTypes:        t.Types,
		ReturnType:        enumType,
		IsBuiltin:         false,
	}
}

func (t *VeniceAnyType) veniceType()        {}
func (t *VeniceBooleanType) veniceType()    {}
func (t *VeniceCharacterType) veniceType()  {}
func (t *VeniceClassType) veniceType()      {}
func (t *VeniceEnumType) veniceType()       {}
func (t *VeniceFunctionType) veniceType()   {}
func (t *VeniceIntegerType) veniceType()    {}
func (t *VeniceListType) veniceType()       {}
func (t *VeniceMapType) veniceType()        {}
func (t *VeniceModuleType) veniceType()     {}
func (t *VeniceRealNumberType) veniceType() {}
func (t *VeniceStringType) veniceType()     {}
func (t *VeniceSymbolType) veniceType()     {}
func (t *VeniceTupleType) veniceType()      {}
func (t *VeniceUnionType) veniceType()      {}
