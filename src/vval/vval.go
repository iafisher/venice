/**
 * Data structures for the representation of Venice values by the virtual machine.
 */
package vval

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
	ClassName string
	Values    []VeniceValue
}

type VeniceEnumObject struct {
	Label  string
	Values []VeniceValue
}

type VeniceFunctionObject struct {
	Name      string
	IsBuiltin bool
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

func VeniceOptionalOf(value VeniceValue) *VeniceEnumObject {
	return &VeniceEnumObject{
		Label:  "Some",
		Values: []VeniceValue{value},
	}
}

var VENICE_OPTIONAL_NONE = &VeniceEnumObject{Label: "None", Values: nil}

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
	return fmt.Sprintf("%q", v.Value)
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

func (v *VeniceFunctionObject) String() string {
	return fmt.Sprintf("<function %q>", v.Name)
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

func (v *VeniceBoolean) Equals(otherAny VeniceValue) bool {
	switch other := otherAny.(type) {
	case *VeniceBoolean:
		return v.Value == other.Value
	default:
		return false
	}
}

func (v *VeniceCharacter) Equals(otherAny VeniceValue) bool {
	switch other := otherAny.(type) {
	case *VeniceCharacter:
		return v.Value == other.Value
	default:
		return false
	}
}

func (v *VeniceClassObject) Equals(otherAny VeniceValue) bool {
	other, ok := otherAny.(*VeniceClassObject)
	if !ok {
		return false
	}

	// We don't check `v.Name == other.Name` because the type-checker already guarantees
	// they will be of the same type.

	if len(v.Values) != len(other.Values) {
		return false
	}

	for i := 0; i < len(v.Values); i++ {
		if !v.Values[i].Equals(other.Values[i]) {
			return false
		}
	}

	return true
}

func (v *VeniceEnumObject) Equals(otherAny VeniceValue) bool {
	other, ok := otherAny.(*VeniceEnumObject)
	if !ok {
		return false
	}

	if v.Label != other.Label {
		return false
	}

	if len(v.Values) != len(other.Values) {
		return false
	}

	for i := 0; i < len(v.Values); i++ {
		if !v.Values[i].Equals(other.Values[i]) {
			return false
		}
	}

	return true
}

func (v *VeniceFunctionObject) Equals(otherAny VeniceValue) bool {
	return false
}

func (v *VeniceInteger) Equals(otherAny VeniceValue) bool {
	other, ok := otherAny.(*VeniceInteger)
	if !ok {
		return false
	}
	return v.Value == other.Value
}

func (v *VeniceList) Equals(otherAny VeniceValue) bool {
	other, ok := otherAny.(*VeniceList)
	if !ok {
		return false
	}

	if len(v.Values) != len(other.Values) {
		return false
	}

	for i := 0; i < len(v.Values); i++ {
		if !v.Values[i].Equals(other.Values[i]) {
			return false
		}
	}

	return true
}

func (v *VeniceListIterator) Equals(otherAny VeniceValue) bool {
	return false
}

func (v *VeniceMap) Equals(otherAny VeniceValue) bool {
	other, ok := otherAny.(*VeniceMap)
	if !ok {
		return false
	}

	if len(v.Pairs) != len(other.Pairs) {
		return false
	}

	for _, pair := range v.Pairs {
		found := false
		for _, otherPair := range other.Pairs {
			if pair.Key.Equals(otherPair.Key) {
				if pair.Value.Equals(otherPair.Value) {
					found = true
					break
				} else {
					return false
				}
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func (v *VeniceMapIterator) Equals(otherAny VeniceValue) bool {
	return false
}

func (v *VeniceString) Equals(otherAny VeniceValue) bool {
	other, ok := otherAny.(*VeniceString)
	if !ok {
		return false
	}
	return v.Value == other.Value
}

func (v *VeniceTuple) Equals(otherAny VeniceValue) bool {
	other, ok := otherAny.(*VeniceTuple)
	if !ok {
		return false
	}

	if len(v.Values) != len(other.Values) {
		return false
	}

	for i := 0; i < len(v.Values); i++ {
		if !v.Values[i].Equals(other.Values[i]) {
			return false
		}
	}

	return true
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

/**
 * Miscellaneous methods
 */

func (v *VeniceMap) Entries() VeniceValue {
	values := []VeniceValue{}
	for _, pair := range v.Pairs {
		values = append(
			values,
			&VeniceTuple{
				[]VeniceValue{
					pair.Key,
					pair.Value,
				},
			},
		)
	}
	return &VeniceList{values}
}

func (v *VeniceMap) Get(key VeniceValue) VeniceValue {
	for _, pair := range v.Pairs {
		if pair.Key.Equals(key) {
			return pair.Value
		}
	}

	return nil
}

func (v *VeniceMap) Keys() VeniceValue {
	values := []VeniceValue{}
	for _, pair := range v.Pairs {
		values = append(values, pair.Key)
	}
	return &VeniceList{values}
}

func (v *VeniceMap) Put(key VeniceValue, value VeniceValue) {
	for _, pair := range v.Pairs {
		if pair.Key.Equals(key) {
			pair.Value = value
			return
		}
	}

	v.Pairs = append(v.Pairs, &VeniceMapPair{Key: key, Value: value})
}

func (v *VeniceMap) Remove(key VeniceValue) {
	for i, pair := range v.Pairs {
		if pair.Key.Equals(key) {
			v.Pairs = append(v.Pairs[:i], v.Pairs[i+1:]...)
			return
		}
	}
}

func (v *VeniceMap) Values() VeniceValue {
	values := []VeniceValue{}
	for _, pair := range v.Pairs {
		values = append(values, pair.Value)
	}
	return &VeniceList{values}
}

func (v *VeniceBoolean) veniceValue()        {}
func (v *VeniceCharacter) veniceValue()      {}
func (v *VeniceClassObject) veniceValue()    {}
func (v *VeniceEnumObject) veniceValue()     {}
func (v *VeniceFunctionObject) veniceValue() {}
func (v *VeniceInteger) veniceValue()        {}
func (v *VeniceList) veniceValue()           {}
func (v *VeniceListIterator) veniceValue()   {}
func (v *VeniceMap) veniceValue()            {}
func (v *VeniceMapIterator) veniceValue()    {}
func (v *VeniceString) veniceValue()         {}
func (v *VeniceTuple) veniceValue()          {}
