package main

import (
	"fmt"
	"strings"
)

type VeniceType interface {
	fmt.Stringer
	veniceType()
	SubstituteGenerics(labels []string, concreteTypes []VeniceType) VeniceType
}

type VeniceGenericType struct {
	Parameters  []string
	GenericType VeniceType
}

func (t *VeniceGenericType) veniceType() {}

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

type VeniceClassType struct {
	Fields []*VeniceClassField
}

func (t *VeniceClassType) veniceType() {}

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

type VeniceClassField struct {
	Name      string
	Public    bool
	FieldType VeniceType
}

type VeniceEnumType struct {
	Cases []*VeniceCaseType
}

func (t *VeniceEnumType) veniceType() {}

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

type VeniceCaseType struct {
	Label string
	Types []VeniceType
}

type VeniceFunctionType struct {
	ParamTypes []VeniceType
	ReturnType VeniceType
}

func (t *VeniceFunctionType) veniceType() {}

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

func (t *VeniceFunctionType) SubstituteGenerics(labels []string, concreteTypes []VeniceType) VeniceType {
	paramTypes := []VeniceType{}
	for _, paramType := range t.ParamTypes {
		paramTypes = append(paramTypes, paramType.SubstituteGenerics(labels, concreteTypes))
	}
	return &VeniceFunctionType{
		paramTypes,
		t.ReturnType.SubstituteGenerics(labels, concreteTypes),
	}
}

type VeniceMapType struct {
	KeyType   VeniceType
	ValueType VeniceType
}

func (t *VeniceMapType) veniceType() {}

func (t *VeniceMapType) String() string {
	return fmt.Sprintf("map<%s, %s>", t.KeyType.String(), t.ValueType.String())
}

func (t *VeniceMapType) SubstituteGenerics(labels []string, concreteTypes []VeniceType) VeniceType {
	return &VeniceMapType{
		t.KeyType.SubstituteGenerics(labels, concreteTypes),
		t.ValueType.SubstituteGenerics(labels, concreteTypes),
	}
}

type VeniceListType struct {
	ItemType VeniceType
}

func (t *VeniceListType) veniceType() {}

func (t *VeniceListType) String() string {
	return fmt.Sprintf("list<%s>", t.ItemType.String())
}

func (t *VeniceListType) SubstituteGenerics(labels []string, concreteTypes []VeniceType) VeniceType {
	return &VeniceListType{
		t.ItemType.SubstituteGenerics(labels, concreteTypes),
	}
}

type VeniceAtomicType struct {
	Type string
}

func (t *VeniceAtomicType) veniceType() {}

func (t *VeniceAtomicType) String() string {
	return t.Type
}

func (t *VeniceAtomicType) SubstituteGenerics(labels []string, concreteTypes []VeniceType) VeniceType {
	return &VeniceAtomicType{t.Type}
}

type VeniceGenericParameterType struct {
	Label string
}

func (t *VeniceGenericParameterType) veniceType() {}

func (t *VeniceGenericParameterType) String() string {
	return t.Label
}

func (t *VeniceGenericParameterType) SubstituteGenerics(labels []string, concreteTypes []VeniceType) VeniceType {
	for i, label := range labels {
		if label == t.Label {
			return concreteTypes[i]
		}
	}
	return &VeniceGenericParameterType{t.Label}
}
