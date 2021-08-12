package main

import (
	"fmt"
	"strings"
)

type VeniceValue interface {
	veniceValue()
	Serialize() string
	Equals(v VeniceValue) bool
}

type VeniceClassObject struct {
	Values []VeniceValue
}

func (v *VeniceClassObject) veniceValue() {}

func (v *VeniceClassObject) Serialize() string {
	var sb strings.Builder
	sb.WriteString("<object ")
	for i, value := range v.Values {
		sb.WriteString(value.Serialize())
		if i != len(v.Values)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte('>')
	return sb.String()
}

func (v *VeniceClassObject) Equals(otherInterface VeniceValue) bool {
	// TODO(2021-08-08): Implement.
	return false
}

type VeniceEnumObject struct {
	Label  string
	Values []VeniceValue
}

func (v *VeniceEnumObject) veniceValue() {}

func (v *VeniceEnumObject) Serialize() string {
	if len(v.Values) == 0 {
		return v.Label
	}

	var sb strings.Builder
	sb.WriteString(v.Label)
	sb.WriteByte('(')
	for i, value := range v.Values {
		sb.WriteString(value.Serialize())
		if i != len(v.Values)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte(')')
	return sb.String()
}

func (v *VeniceEnumObject) Equals(otherInterface VeniceValue) bool {
	// TODO(2021-08-09): Implement.
	return false
}

type VeniceList struct {
	Values []VeniceValue
}

func (v *VeniceList) veniceValue() {}

func (v *VeniceList) Serialize() string {
	var sb strings.Builder
	sb.WriteByte('[')
	for i, value := range v.Values {
		sb.WriteString(value.Serialize())
		if i != len(v.Values)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte(']')
	return sb.String()
}

func (v *VeniceList) Equals(otherInterface VeniceValue) bool {
	switch other := otherInterface.(type) {
	case *VeniceList:
		if len(v.Values) != len(other.Values) {
			return false
		}

		for i := 0; i < len(v.Values); i++ {
			if !v.Values[i].Equals(other.Values[i]) {
				return false
			}
		}

		return true
	default:
		return false
	}
}

type VeniceTuple struct {
	Values []VeniceValue
}

func (v *VeniceTuple) veniceValue() {}

func (v *VeniceTuple) Serialize() string {
	var sb strings.Builder
	sb.WriteByte('(')
	for i, value := range v.Values {
		sb.WriteString(value.Serialize())
		if i != len(v.Values)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte(')')
	return sb.String()
}

func (v *VeniceTuple) Equals(otherInterface VeniceValue) bool {
	switch other := otherInterface.(type) {
	case *VeniceTuple:
		if len(v.Values) != len(other.Values) {
			return false
		}

		for i := 0; i < len(v.Values); i++ {
			if !v.Values[i].Equals(other.Values[i]) {
				return false
			}
		}

		return true
	default:
		return false
	}
}

type VeniceMap struct {
	Pairs []*VeniceMapPair
}

func (v *VeniceMap) veniceValue() {}

func (v *VeniceMap) Serialize() string {
	var sb strings.Builder
	sb.WriteByte('{')
	for i, pair := range v.Pairs {
		sb.WriteString(pair.Key.Serialize())
		sb.WriteString(": ")
		sb.WriteString(pair.Value.Serialize())
		if i != len(v.Pairs)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte('}')
	return sb.String()
}

func (v *VeniceMap) Equals(otherInterface VeniceValue) bool {
	// TODO(2021-08-31): Implement.
	return false
}

type VeniceMapPair struct {
	Key   VeniceValue
	Value VeniceValue
}

type VeniceInteger struct {
	Value int
}

func (v *VeniceInteger) veniceValue() {}

func (v *VeniceInteger) Serialize() string {
	return fmt.Sprintf("%d", v.Value)
}

func (v *VeniceInteger) Equals(otherInterface VeniceValue) bool {
	switch other := otherInterface.(type) {
	case *VeniceInteger:
		return v.Value == other.Value
	default:
		return false
	}
}

type VeniceString struct {
	Value string
}

func (v *VeniceString) veniceValue() {}

func (v *VeniceString) Serialize() string {
	return fmt.Sprintf("%q", v.Value)
}

func (v *VeniceString) Equals(otherInterface VeniceValue) bool {
	switch other := otherInterface.(type) {
	case *VeniceString:
		return v.Value == other.Value
	default:
		return false
	}
}

type VeniceBoolean struct {
	Value bool
}

func (v *VeniceBoolean) veniceValue() {}

func (v *VeniceBoolean) Serialize() string {
	if v.Value {
		return "true"
	} else {
		return "false"
	}
}

func (v *VeniceBoolean) Equals(otherInterface VeniceValue) bool {
	switch other := otherInterface.(type) {
	case *VeniceBoolean:
		return v.Value == other.Value
	default:
		return false
	}
}

type VeniceCharacter struct {
	Value byte
}

func (v *VeniceCharacter) veniceValue() {}

func (v *VeniceCharacter) Serialize() string {
	return fmt.Sprintf("'%c'", v.Value)
}

func (v *VeniceCharacter) Equals(otherInterface VeniceValue) bool {
	switch other := otherInterface.(type) {
	case *VeniceCharacter:
		return v.Value == other.Value
	default:
		return false
	}
}

type VeniceFunction struct {
	Params []string
	Body   []*Bytecode
}

func (v *VeniceFunction) veniceValue() {}

func (v *VeniceFunction) Serialize() string {
	return "<function object>"
}

func (v *VeniceFunction) Equals(otherInterface VeniceValue) bool {
	return false
}

type VeniceIterator interface {
	VeniceValue
	Next() []VeniceValue
}

type VeniceListIterator struct {
	List  *VeniceList
	Index int
}

func (v *VeniceListIterator) veniceValue() {}

func (v *VeniceListIterator) Serialize() string {
	return "<list iterator>"
}

func (v *VeniceListIterator) Equals(otherInterface VeniceValue) bool {
	return false
}

func (v *VeniceListIterator) Next() []VeniceValue {
	if v.Index == len(v.List.Values) {
		return nil
	} else {
		r := v.List.Values[v.Index]
		v.Index++
		return []VeniceValue{r}
	}
}

type VeniceMapIterator struct {
	Map   *VeniceMap
	Index int
}

func (v *VeniceMapIterator) veniceValue() {}

func (v *VeniceMapIterator) Serialize() string {
	return "<map iterator>"
}

func (v *VeniceMapIterator) Equals(otherInterface VeniceValue) bool {
	return false
}

func (v *VeniceMapIterator) Next() []VeniceValue {
	if v.Index == len(v.Map.Pairs) {
		return nil
	} else {
		r := v.Map.Pairs[v.Index]
		v.Index++
		return []VeniceValue{r.Key, r.Value}
	}
}
