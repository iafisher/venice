/**
 * Data structures for the representation of Venice values by the virtual machine.
 */
package main

import (
	"fmt"
	"strings"
)

type VeniceValue interface {
	fmt.Stringer
	veniceValue()
	Equals(v VeniceValue) bool
}

/**
 * Compound types
 */

type VeniceClassObject struct {
	Values []VeniceValue
}

type VeniceEnumObject struct {
	Label  string
	Values []VeniceValue
}

type VeniceFunction struct {
	Params []string
	Body   []*Bytecode
}

type VeniceList struct {
	Values []VeniceValue
}

type VeniceMap struct {
	Pairs []*VeniceMapPair
}

// Helper struct - does not implement VeniceValue
type VeniceMapPair struct {
	Key   VeniceValue
	Value VeniceValue
}

type VeniceTuple struct {
	Values []VeniceValue
}

/**
 * Iterator types
 */

type VeniceIterator interface {
	VeniceValue
	Next() []VeniceValue
}

type VeniceMapIterator struct {
	Map   *VeniceMap
	Index int
}

type VeniceListIterator struct {
	List  *VeniceList
	Index int
}

/**
 * Atomic types
 */

type VeniceBoolean struct {
	Value bool
}

type VeniceCharacter struct {
	Value byte
}

type VeniceInteger struct {
	Value int
}

type VeniceString struct {
	Value string
}

/**
 * String() implementations
 */

func (v *VeniceBoolean) String() string {
	if v.Value {
		return "true"
	} else {
		return "false"
	}
}

func (v *VeniceCharacter) String() string {
	// TODO(2021-08-13): Does this handle backslash escapes correctly?
	return fmt.Sprintf("'%c'", v.Value)
}

func (v *VeniceClassObject) String() string {
	var sb strings.Builder
	sb.WriteString("<object ")
	for i, value := range v.Values {
		sb.WriteString(value.String())
		if i != len(v.Values)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte('>')
	return sb.String()
}

func (v *VeniceEnumObject) String() string {
	if len(v.Values) == 0 {
		return v.Label
	}

	var sb strings.Builder
	sb.WriteString(v.Label)
	sb.WriteByte('(')
	for i, value := range v.Values {
		sb.WriteString(value.String())
		if i != len(v.Values)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte(')')
	return sb.String()
}

func (v *VeniceFunction) String() string {
	return "<function object>"
}

func (v *VeniceInteger) String() string {
	return fmt.Sprintf("%d", v.Value)
}

func (v *VeniceList) String() string {
	var sb strings.Builder
	sb.WriteByte('[')
	for i, value := range v.Values {
		sb.WriteString(value.String())
		if i != len(v.Values)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte(']')
	return sb.String()
}

func (v *VeniceListIterator) String() string {
	return "<list iterator>"
}

func (v *VeniceMap) String() string {
	var sb strings.Builder
	sb.WriteByte('{')
	for i, pair := range v.Pairs {
		sb.WriteString(pair.Key.String())
		sb.WriteString(": ")
		sb.WriteString(pair.Value.String())
		if i != len(v.Pairs)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte('}')
	return sb.String()
}

func (v *VeniceMapIterator) String() string {
	return "<map iterator>"
}

func (v *VeniceString) String() string {
	return fmt.Sprintf("%q", v.Value)
}

func (v *VeniceTuple) String() string {
	var sb strings.Builder
	sb.WriteByte('(')
	for i, value := range v.Values {
		sb.WriteString(value.String())
		if i != len(v.Values)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte(')')
	return sb.String()
}

/**
 * Equals() implementations
 */

func (v *VeniceBoolean) Equals(otherInterface VeniceValue) bool {
	switch other := otherInterface.(type) {
	case *VeniceBoolean:
		return v.Value == other.Value
	default:
		return false
	}
}

func (v *VeniceCharacter) Equals(otherInterface VeniceValue) bool {
	switch other := otherInterface.(type) {
	case *VeniceCharacter:
		return v.Value == other.Value
	default:
		return false
	}
}

func (v *VeniceClassObject) Equals(otherInterface VeniceValue) bool {
	// TODO(2021-08-08): Implement.
	return false
}

func (v *VeniceEnumObject) Equals(otherInterface VeniceValue) bool {
	// TODO(2021-08-09): Implement.
	return false
}

func (v *VeniceFunction) Equals(otherInterface VeniceValue) bool {
	return false
}

func (v *VeniceInteger) Equals(otherInterface VeniceValue) bool {
	switch other := otherInterface.(type) {
	case *VeniceInteger:
		return v.Value == other.Value
	default:
		return false
	}
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

func (v *VeniceListIterator) Equals(otherInterface VeniceValue) bool {
	return false
}

func (v *VeniceMap) Equals(otherInterface VeniceValue) bool {
	// TODO(2021-08-31): Implement.
	return false
}

func (v *VeniceMapIterator) Equals(otherInterface VeniceValue) bool {
	return false
}

func (v *VeniceString) Equals(otherInterface VeniceValue) bool {
	switch other := otherInterface.(type) {
	case *VeniceString:
		return v.Value == other.Value
	default:
		return false
	}
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

/**
 * Next() implementations for iterators
 */

func (v *VeniceListIterator) Next() []VeniceValue {
	if v.Index == len(v.List.Values) {
		return nil
	} else {
		r := v.List.Values[v.Index]
		v.Index++
		return []VeniceValue{r}
	}
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

func (v *VeniceBoolean) veniceValue()      {}
func (v *VeniceCharacter) veniceValue()    {}
func (v *VeniceClassObject) veniceValue()  {}
func (v *VeniceEnumObject) veniceValue()   {}
func (v *VeniceFunction) veniceValue()     {}
func (v *VeniceInteger) veniceValue()      {}
func (v *VeniceListIterator) veniceValue() {}
func (v *VeniceMap) veniceValue()          {}
func (v *VeniceMapIterator) veniceValue()  {}
func (v *VeniceList) veniceValue()         {}
func (v *VeniceString) veniceValue()       {}
func (v *VeniceTuple) veniceValue()        {}
