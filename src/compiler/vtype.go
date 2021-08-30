package compiler

import (
	"fmt"
	"strings"
)

type VeniceType interface {
	fmt.Stringer
	veniceType()
	MatchGenerics(genericParameterMap map[string]VeniceType, concreteType VeniceType) error
	SubstituteGenerics(genericParameterMap map[string]VeniceType) VeniceType
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
	return VENICE_TYPE_OPTIONAL.SubstituteGenerics(map[string]VeniceType{"T": concreteType})
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
 * SubstituteGenerics() implementations
 *
 * - Compound types recursively call SubstituteGenerics() on their constituent types.
 * - Atomic types do nothing.
 * - Generic parameter types substitute the concrete type if the parameter matches.
 */

func (t *VeniceAnyType) SubstituteGenerics(
	genericParameterMap map[string]VeniceType,
) VeniceType {
	return &VeniceAnyType{}
}

func (t *VeniceBooleanType) SubstituteGenerics(
	genericParameterMap map[string]VeniceType,
) VeniceType {
	return &VeniceBooleanType{}
}

func (t *VeniceCharacterType) SubstituteGenerics(
	genericParameterMap map[string]VeniceType,
) VeniceType {
	return &VeniceCharacterType{}
}

func (t *VeniceClassType) SubstituteGenerics(
	genericParameterMap map[string]VeniceType,
) VeniceType {
	fields := make([]*VeniceClassField, 0, len(t.Fields))
	for _, field := range t.Fields {
		fields = append(fields, &VeniceClassField{
			field.Name,
			field.Public,
			field.FieldType.SubstituteGenerics(genericParameterMap),
		})
	}
	return &VeniceClassType{Name: t.Name, Fields: fields, Methods: t.Methods}
}

func (t *VeniceEnumType) SubstituteGenerics(
	genericParameterMap map[string]VeniceType,
) VeniceType {
	cases := make([]*VeniceCaseType, 0, len(t.Cases))
	for _, enumCase := range t.Cases {
		caseTypes := make([]VeniceType, 0, len(enumCase.Types))
		for _, caseType := range enumCase.Types {
			caseTypes = append(caseTypes, caseType.SubstituteGenerics(genericParameterMap))
		}
		cases = append(cases, &VeniceCaseType{enumCase.Label, caseTypes})
	}
	return &VeniceEnumType{Name: t.Name, Cases: cases}
}

func (t *VeniceFunctionType) SubstituteGenerics(
	genericParameterMap map[string]VeniceType,
) VeniceType {
	paramTypes := make([]VeniceType, 0, len(t.ParamTypes))
	for _, paramType := range t.ParamTypes {
		paramTypes = append(paramTypes, paramType.SubstituteGenerics(genericParameterMap))
	}
	return &VeniceFunctionType{
		Name:       t.Name,
		ParamTypes: paramTypes,
		ReturnType: t.ReturnType.SubstituteGenerics(genericParameterMap),
		IsBuiltin:  t.IsBuiltin,
	}
}

func (t *VeniceIntegerType) SubstituteGenerics(
	genericParameterMap map[string]VeniceType,
) VeniceType {
	return &VeniceIntegerType{}
}

func (t *VeniceListType) SubstituteGenerics(
	genericParameterMap map[string]VeniceType,
) VeniceType {
	return &VeniceListType{
		t.ItemType.SubstituteGenerics(genericParameterMap),
	}
}

func (t *VeniceMapType) SubstituteGenerics(
	genericParameterMap map[string]VeniceType,
) VeniceType {
	return &VeniceMapType{
		t.KeyType.SubstituteGenerics(genericParameterMap),
		t.ValueType.SubstituteGenerics(genericParameterMap),
	}
}

func (t *VeniceModuleType) SubstituteGenerics(
	genericParameterMap map[string]VeniceType,
) VeniceType {
	// TODO(2021-08-24): Is this right?
	return t
}

func (t *VeniceRealNumberType) SubstituteGenerics(
	genericParameterMap map[string]VeniceType,
) VeniceType {
	return &VeniceRealNumberType{}
}

func (t *VeniceStringType) SubstituteGenerics(
	genericParameterMap map[string]VeniceType,
) VeniceType {
	return &VeniceStringType{}
}

func (t *VeniceSymbolType) SubstituteGenerics(
	genericParameterMap map[string]VeniceType,
) VeniceType {
	if concreteType, ok := genericParameterMap[t.Label]; ok {
		return concreteType
	} else {
		return &VeniceSymbolType{t.Label}
	}
}

func (t *VeniceTupleType) SubstituteGenerics(
	genericParameterMap map[string]VeniceType,
) VeniceType {
	newItemTypes := make([]VeniceType, 0, len(t.ItemTypes))
	for _, itemType := range t.ItemTypes {
		newItemTypes = append(newItemTypes, itemType.SubstituteGenerics(genericParameterMap))
	}
	return &VeniceTupleType{newItemTypes}
}

func (t *VeniceUnionType) SubstituteGenerics(
	genericParameterMap map[string]VeniceType,
) VeniceType {
	newTypes := make([]VeniceType, 0, len(t.Types))
	for _, subType := range t.Types {
		newTypes = append(newTypes, subType.SubstituteGenerics(genericParameterMap))
	}
	return &VeniceUnionType{newTypes}
}

/**
 * MatchGenerics() implementations
 */

func (t *VeniceAnyType) MatchGenerics(
	genericParameterMap map[string]VeniceType, concreteType VeniceType,
) error {
	return nil
}

func (t *VeniceBooleanType) MatchGenerics(
	genericParameterMap map[string]VeniceType, concreteType VeniceType,
) error {
	return nil
}

func (t *VeniceCharacterType) MatchGenerics(
	genericParameterMap map[string]VeniceType, concreteType VeniceType,
) error {
	return nil
}

func (t *VeniceClassType) MatchGenerics(
	genericParameterMap map[string]VeniceType, concreteType VeniceType,
) error {
	// TODO(2021-08-20): Is this right?
	return nil
}

func (t *VeniceEnumType) MatchGenerics(
	genericParameterMap map[string]VeniceType, concreteType VeniceType,
) error {
	if enumType, ok := concreteType.(*VeniceEnumType); ok {
		for i, enumCase := range t.Cases {
			if i >= len(enumType.Cases) {
				break
			}

			for j, caseType := range enumCase.Types {
				if j >= len(enumType.Cases[i].Types) {
					break
				}

				caseType.MatchGenerics(genericParameterMap, enumType.Cases[i].Types[j])
			}
		}
	}
	return nil
}

func (t *VeniceFunctionType) MatchGenerics(
	genericParameterMap map[string]VeniceType, concreteType VeniceType,
) error {
	if functionType, ok := concreteType.(*VeniceFunctionType); ok {
		for i, paramType := range t.ParamTypes {
			if i >= len(functionType.ParamTypes) {
				break
			}
			paramType.MatchGenerics(genericParameterMap, functionType.ParamTypes[i])
		}
		t.ReturnType.MatchGenerics(genericParameterMap, functionType.ReturnType)
	}
	return nil
}

func (t *VeniceIntegerType) MatchGenerics(
	genericParameterMap map[string]VeniceType, concreteType VeniceType,
) error {
	return nil
}

func (t *VeniceListType) MatchGenerics(
	genericParameterMap map[string]VeniceType, concreteType VeniceType,
) error {
	if listType, ok := concreteType.(*VeniceListType); ok {
		t.ItemType.MatchGenerics(genericParameterMap, listType.ItemType)
	}
	return nil
}

func (t *VeniceMapType) MatchGenerics(
	genericParameterMap map[string]VeniceType, concreteType VeniceType,
) error {
	if mapType, ok := concreteType.(*VeniceMapType); ok {
		t.KeyType.MatchGenerics(genericParameterMap, mapType.KeyType)
		t.ValueType.MatchGenerics(genericParameterMap, mapType.ValueType)
	}
	return nil
}

func (t *VeniceModuleType) MatchGenerics(
	genericParameterMap map[string]VeniceType, concreteType VeniceType,
) error {
	// TODO(2021-08-24): Is this right?
	return nil
}

func (t *VeniceRealNumberType) MatchGenerics(
	genericParameterMap map[string]VeniceType, concreteType VeniceType,
) error {
	return nil
}

func (t *VeniceStringType) MatchGenerics(
	genericParameterMap map[string]VeniceType, concreteType VeniceType,
) error {
	return nil
}

func (t *VeniceSymbolType) MatchGenerics(
	genericParameterMap map[string]VeniceType, concreteType VeniceType,
) error {
	// TODO(2021-08-20): Check for existing type.
	genericParameterMap[t.Label] = concreteType
	return nil
}

func (t *VeniceTupleType) MatchGenerics(
	genericParameterMap map[string]VeniceType, concreteType VeniceType,
) error {
	if tupleType, ok := concreteType.(*VeniceTupleType); ok {
		for i, itemType := range t.ItemTypes {
			if i >= len(tupleType.ItemTypes) {
				break
			}
			itemType.MatchGenerics(genericParameterMap, tupleType.ItemTypes[i])
		}
	}
	return nil
}

func (t *VeniceUnionType) MatchGenerics(
	genericParameterMap map[string]VeniceType, concreteType VeniceType,
) error {
	if unionType, ok := concreteType.(*VeniceUnionType); ok {
		for i, subType := range t.Types {
			if i >= len(unionType.Types) {
				break
			}
			subType.MatchGenerics(genericParameterMap, unionType.Types[i])
		}
	}
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
