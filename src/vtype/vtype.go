package vtype

import (
	"fmt"
	"strings"
)

type VeniceType interface {
	fmt.Stringer
	veniceType()
	SubstituteGenerics(labels []string, concreteTypes []VeniceType) VeniceType
}

/**
 * Compound types
 */

type VeniceAtomicType struct {
	Type string
}

type VeniceClassType struct {
	Fields []*VeniceClassField
}

// Helper struct - does not implement VeniceType
type VeniceClassField struct {
	Name      string
	Public    bool
	FieldType VeniceType
}

type VeniceEnumType struct {
	Cases []*VeniceCaseType
}

// Helper struct - does not implement VeniceType
type VeniceCaseType struct {
	Label string
	Types []VeniceType
}

type VeniceFunctionType struct {
	ParamTypes []VeniceType
	ReturnType VeniceType
	IsBuiltin  bool
}

type VeniceGenericParameterType struct {
	Label string
}

type VeniceGenericType struct {
	Parameters  []string
	GenericType VeniceType
}

type VeniceListType struct {
	ItemType VeniceType
}

type VeniceMapType struct {
	KeyType   VeniceType
	ValueType VeniceType
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

const (
	VENICE_TYPE_ANY_LABEL       = "any"
	VENICE_TYPE_BOOLEAN_LABEL   = "bool"
	VENICE_TYPE_CHARACTER_LABEL = "char"
	VENICE_TYPE_INTEGER_LABEL   = "int"
	VENICE_TYPE_STRING_LABEL    = "string"
)

var (
	VENICE_TYPE_ANY       = &VeniceAtomicType{VENICE_TYPE_ANY_LABEL}
	VENICE_TYPE_BOOLEAN   = &VeniceAtomicType{VENICE_TYPE_BOOLEAN_LABEL}
	VENICE_TYPE_CHARACTER = &VeniceAtomicType{VENICE_TYPE_CHARACTER_LABEL}
	VENICE_TYPE_INTEGER   = &VeniceAtomicType{VENICE_TYPE_INTEGER_LABEL}
	VENICE_TYPE_STRING    = &VeniceAtomicType{VENICE_TYPE_STRING_LABEL}
)

/**
 * String() implementations
 */

func (t *VeniceAtomicType) String() string {
	return t.Type
}

func (t *VeniceClassType) String() string {
	var sb strings.Builder
	sb.WriteString("class(")
	for i, field := range t.Fields {
		if field.Public {
			sb.WriteString("public ")
		} else {
			sb.WriteString("private ")
		}

		sb.WriteString(field.Name)
		sb.WriteString(": ")
		sb.WriteString(field.FieldType.String())

		if i != len(t.Fields)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte(')')
	return sb.String()
}

func (t *VeniceEnumType) String() string {
	var sb strings.Builder
	sb.WriteString("enum(")
	for i, enumCase := range t.Cases {
		sb.WriteString(enumCase.Label)
		sb.WriteByte('(')
		for j, caseType := range enumCase.Types {
			sb.WriteString(caseType.String())
			if j != len(enumCase.Types)-1 {
				sb.WriteString(", ")
			}
		}
		sb.WriteByte(')')

		if i != len(t.Cases)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte(')')
	return sb.String()
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

func (t *VeniceGenericParameterType) String() string {
	return t.Label
}

func (t *VeniceGenericType) String() string {
	var sb strings.Builder
	sb.WriteString("type<")
	for i, parameter := range t.Parameters {
		sb.WriteString(parameter)
		if i != len(t.Parameters)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte('>')
	sb.WriteByte('(')
	sb.WriteString(t.GenericType.String())
	sb.WriteByte(')')
	return sb.String()
}

func (t *VeniceListType) String() string {
	return fmt.Sprintf("list<%s>", t.ItemType.String())
}

func (t *VeniceMapType) String() string {
	return fmt.Sprintf("map<%s, %s>", t.KeyType.String(), t.ValueType.String())
}

func (t *VeniceTupleType) String() string {
	var sb strings.Builder
	sb.WriteString("tuple<")
	for i, itemType := range t.ItemTypes {
		sb.WriteString(itemType.String())
		if i != len(t.ItemTypes)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte('>')
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

func (t *VeniceAtomicType) SubstituteGenerics(labels []string, concreteTypes []VeniceType) VeniceType {
	return &VeniceAtomicType{t.Type}
}

func (t *VeniceClassType) SubstituteGenerics(labels []string, concreteTypes []VeniceType) VeniceType {
	fields := []*VeniceClassField{}
	for _, field := range t.Fields {
		fields = append(fields, &VeniceClassField{
			field.Name,
			field.Public,
			field.FieldType.SubstituteGenerics(labels, concreteTypes),
		})
	}
	return &VeniceClassType{fields}
}

func (t *VeniceEnumType) SubstituteGenerics(labels []string, concreteTypes []VeniceType) VeniceType {
	cases := []*VeniceCaseType{}
	for _, enumCase := range t.Cases {
		caseTypes := []VeniceType{}
		for _, caseType := range enumCase.Types {
			caseTypes = append(caseTypes, caseType.SubstituteGenerics(labels, concreteTypes))
		}
		cases = append(cases, &VeniceCaseType{enumCase.Label, caseTypes})
	}
	return &VeniceEnumType{cases}
}

func (t *VeniceFunctionType) SubstituteGenerics(labels []string, concreteTypes []VeniceType) VeniceType {
	paramTypes := []VeniceType{}
	for _, paramType := range t.ParamTypes {
		paramTypes = append(paramTypes, paramType.SubstituteGenerics(labels, concreteTypes))
	}
	return &VeniceFunctionType{
		paramTypes,
		t.ReturnType.SubstituteGenerics(labels, concreteTypes),
		t.IsBuiltin,
	}
}

func (t *VeniceGenericParameterType) SubstituteGenerics(labels []string, concreteTypes []VeniceType) VeniceType {
	for i, label := range labels {
		if label == t.Label {
			return concreteTypes[i]
		}
	}
	return &VeniceGenericParameterType{t.Label}
}

func (t *VeniceGenericType) SubstituteGenerics(labels []string, concreteTypes []VeniceType) VeniceType {
	parameters := []string{}
	for _, param := range t.Parameters {
		found := false
		for _, label := range labels {
			if param == label {
				found = true
				break
			}
		}

		if !found {
			parameters = append(parameters, param)
		}
	}

	if len(parameters) == 0 {
		return t.GenericType.SubstituteGenerics(labels, concreteTypes)
	} else {
		return &VeniceGenericType{
			parameters,
			t.GenericType.SubstituteGenerics(labels, concreteTypes),
		}
	}
}

func (t *VeniceListType) SubstituteGenerics(labels []string, concreteTypes []VeniceType) VeniceType {
	return &VeniceListType{
		t.ItemType.SubstituteGenerics(labels, concreteTypes),
	}
}

func (t *VeniceMapType) SubstituteGenerics(labels []string, concreteTypes []VeniceType) VeniceType {
	return &VeniceMapType{
		t.KeyType.SubstituteGenerics(labels, concreteTypes),
		t.ValueType.SubstituteGenerics(labels, concreteTypes),
	}
}

func (t *VeniceTupleType) SubstituteGenerics(labels []string, concreteTypes []VeniceType) VeniceType {
	newItemTypes := []VeniceType{}
	for _, itemType := range t.ItemTypes {
		newItemTypes = append(newItemTypes, itemType.SubstituteGenerics(labels, concreteTypes))
	}
	return &VeniceTupleType{newItemTypes}
}

func (t *VeniceUnionType) SubstituteGenerics(labels []string, concreteTypes []VeniceType) VeniceType {
	newTypes := []VeniceType{}
	for _, subType := range t.Types {
		newTypes = append(newTypes, subType.SubstituteGenerics(labels, concreteTypes))
	}
	return &VeniceUnionType{newTypes}
}

func (t *VeniceAtomicType) veniceType()           {}
func (t *VeniceClassType) veniceType()            {}
func (t *VeniceEnumType) veniceType()             {}
func (t *VeniceFunctionType) veniceType()         {}
func (t *VeniceGenericParameterType) veniceType() {}
func (t *VeniceGenericType) veniceType()          {}
func (t *VeniceListType) veniceType()             {}
func (t *VeniceMapType) veniceType()              {}
func (t *VeniceTupleType) veniceType()            {}
func (t *VeniceUnionType) veniceType()            {}
