package main

import (
	"fmt"
	"strings"
)

type VeniceValue interface {
	veniceValue()
	Serialize() string
	SerializePrintable() string
	Equals(v VeniceValue) bool
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

func (v *VeniceList) SerializePrintable() string {
	var sb strings.Builder
	sb.WriteByte('[')
	for i, value := range v.Values {
		sb.WriteString(value.SerializePrintable())
		if i != len(v.Values)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte(']')
	return sb.String()
}

func (v *VeniceList) Equals(otherUntyped VeniceValue) bool {
	switch other := otherUntyped.(type) {
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

func (v *VeniceMap) SerializePrintable() string {
	var sb strings.Builder
	sb.WriteByte('{')
	for i, pair := range v.Pairs {
		sb.WriteString(pair.Key.SerializePrintable())
		sb.WriteString(": ")
		sb.WriteString(pair.Value.SerializePrintable())
		if i != len(v.Pairs)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte('}')
	return sb.String()
}

func (v *VeniceMap) Equals(otherUntyped VeniceValue) bool {
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

func (v *VeniceInteger) SerializePrintable() string {
	return v.Serialize()
}

func (v *VeniceInteger) Equals(otherUntyped VeniceValue) bool {
	switch other := otherUntyped.(type) {
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

func (v *VeniceString) SerializePrintable() string {
	return v.Value
}

func (v *VeniceString) Equals(otherUntyped VeniceValue) bool {
	switch other := otherUntyped.(type) {
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

func (v *VeniceBoolean) SerializePrintable() string {
	return v.Serialize()
}

func (v *VeniceBoolean) Equals(otherUntyped VeniceValue) bool {
	switch other := otherUntyped.(type) {
	case *VeniceBoolean:
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

func (v *VeniceFunction) SerializePrintable() string {
	return v.Serialize()
}

func (v *VeniceFunction) Equals(otherUntyped VeniceValue) bool {
	return false
}
