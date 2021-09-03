/**
 * Data structures for the representation of Venice values by the virtual machine.
 */
package vm

import (
	"fmt"
	"hash/maphash"
	"strings"
)

type VeniceValue interface {
	fmt.Stringer
	veniceValue()
	Compare(v VeniceValue) bool
	Equals(v VeniceValue) bool
	Hash(h maphash.Hash) uint64
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
	table [256]*VeniceMapChain
	hash  maphash.Hash
	Size  int
}

// Helper struct - does not implement VeniceValue
type VeniceMapChain struct {
	Key   VeniceValue
	Value VeniceValue
	Next  *VeniceMapChain
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
	Map        *VeniceMap
	TableIndex int
	ChainIndex int
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

type VeniceInteger struct {
	Value int
}

type VeniceRealNumber struct {
	Value float64
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
	for i, entry := range v.Entries().Values {
		entryAsTuple := entry.(*VeniceTuple)
		key := entryAsTuple.Values[0]
		value := entryAsTuple.Values[1]

		sb.WriteString(key.String())
		sb.WriteString(": ")
		sb.WriteString(value.String())
		if i != v.Size-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteByte('}')
	return sb.String()
}

func (v *VeniceMapIterator) String() string {
	return "<map iterator>"
}

func (v *VeniceRealNumber) String() string {
	return fmt.Sprintf("%g", v.Value)
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
 * Compare() implementations
 */

func (v *VeniceBoolean) Compare(otherAny VeniceValue) bool {
	switch other := otherAny.(type) {
	case *VeniceBoolean:
		if !v.Value && other.Value {
			return true
		} else {
			return false
		}
	default:
		return false
	}
}

func (v *VeniceClassObject) Compare(otherAny VeniceValue) bool {
	return false
}

func (v *VeniceEnumObject) Compare(otherAny VeniceValue) bool {
	return false
}

func (v *VeniceFunctionObject) Compare(otherAny VeniceValue) bool {
	return false
}

func (v *VeniceInteger) Compare(otherAny VeniceValue) bool {
	switch other := otherAny.(type) {
	case *VeniceInteger:
		return v.Value < other.Value
	case *VeniceRealNumber:
		return float64(v.Value) < other.Value
	default:
		return false
	}
}

func (v *VeniceList) Compare(otherAny VeniceValue) bool {
	return false
}

func (v *VeniceListIterator) Compare(otherAny VeniceValue) bool {
	return false
}

func (v *VeniceMap) Compare(otherAny VeniceValue) bool {
	return false
}

func (v *VeniceMapIterator) Compare(otherAny VeniceValue) bool {
	return false
}

func (v *VeniceRealNumber) Compare(otherAny VeniceValue) bool {
	switch other := otherAny.(type) {
	case *VeniceInteger:
		return v.Value < float64(other.Value)
	case *VeniceRealNumber:
		return v.Value < other.Value
	default:
		return false
	}
}

func (v *VeniceString) Compare(otherAny VeniceValue) bool {
	other, ok := otherAny.(*VeniceString)
	if !ok {
		return false
	}
	return v.Value < other.Value
}

func (v *VeniceTuple) Compare(otherAny VeniceValue) bool {
	return false
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
	switch other := otherAny.(type) {
	case *VeniceInteger:
		return v.Value == other.Value
	case *VeniceRealNumber:
		return float64(v.Value) == other.Value
	default:
		return false
	}
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

	if v.Size != other.Size {
		return false
	}

	for _, entry := range v.Entries().Values {
		entryAsTuple := entry.(*VeniceTuple)
		key := entryAsTuple.Values[0]
		value := entryAsTuple.Values[1]

		if !other.Get(key).Equals(value) {
			return false
		}
	}

	return true
}

func (v *VeniceMapIterator) Equals(otherAny VeniceValue) bool {
	return false
}

func (v *VeniceRealNumber) Equals(otherAny VeniceValue) bool {
	switch other := otherAny.(type) {
	case *VeniceInteger:
		return v.Value == float64(other.Value)
	case *VeniceRealNumber:
		return v.Value == other.Value
	default:
		return false
	}
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
 * Hash() implementations
 */

func (v *VeniceBoolean) Hash(h maphash.Hash) uint64 {
	if v.Value {
		h.WriteByte(1)
	} else {
		h.WriteByte(0)
	}
	return h.Sum64()
}

func (v *VeniceClassObject) Hash(h maphash.Hash) uint64 {
	// TODO(2021-08-30): Handle this properly.
	return 0
}

func (v *VeniceEnumObject) Hash(h maphash.Hash) uint64 {
	// TODO(2021-08-30): Handle this properly.
	return 0
}

func (v *VeniceFunctionObject) Hash(h maphash.Hash) uint64 {
	// TODO(2021-08-30): Handle this properly.
	return 0
}

func (v *VeniceInteger) Hash(h maphash.Hash) uint64 {
	// TODO(2021-08-30): Handle this properly.
	return uint64(v.Value)
}

func (v *VeniceList) Hash(h maphash.Hash) uint64 {
	// TODO(2021-08-30): Is this usage correct?
	for _, value := range v.Values {
		value.Hash(h)
	}
	return h.Sum64()
}

func (v *VeniceListIterator) Hash(h maphash.Hash) uint64 {
	// TODO(2021-08-30): Handle this properly.
	return 0
}

func (v *VeniceMap) Hash(h maphash.Hash) uint64 {
	// TODO(2021-08-30): Handle this properly.
	return 0
}

func (v *VeniceMapIterator) Hash(h maphash.Hash) uint64 {
	// TODO(2021-08-30): Handle this properly.
	return 0
}

func (v *VeniceRealNumber) Hash(h maphash.Hash) uint64 {
	// TODO(2021-08-30): Handle this properly.
	return uint64(v.Value)
}

func (v *VeniceString) Hash(h maphash.Hash) uint64 {
	h.WriteString(v.Value)
	return h.Sum64()
}

func (v *VeniceTuple) Hash(h maphash.Hash) uint64 {
	// TODO(2021-08-30): Is this usage correct?
	for _, value := range v.Values {
		value.Hash(h)
	}
	return h.Sum64()
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
	for v.TableIndex < len(v.Map.table) {
		entry := v.Map.table[v.TableIndex]
		for i := 0; i < v.ChainIndex && entry != nil; i++ {
			entry = entry.Next
		}

		if entry != nil {
			v.ChainIndex++
			return []VeniceValue{entry.Key, entry.Value}
		}

		v.TableIndex++
		v.ChainIndex = 0
	}

	return nil
}

/**
 * Miscellaneous methods
 */

func NewVeniceListIterator(list *VeniceList) *VeniceListIterator {
	return &VeniceListIterator{List: list, Index: 0}
}

func NewVeniceMap() *VeniceMap {
	return &VeniceMap{
		Size: 0,
	}
}

func NewVeniceMapIterator(vmap *VeniceMap) *VeniceMapIterator {
	return &VeniceMapIterator{Map: vmap, TableIndex: 0, ChainIndex: 0}
}

func (v *VeniceList) Copy() *VeniceList {
	copiedValues := make([]VeniceValue, len(v.Values))
	copy(copiedValues, v.Values)
	return &VeniceList{copiedValues}
}

func (v *VeniceMap) Clear() {
	for i := 0; i < len(v.table); i++ {
		v.table[i] = nil
	}
	v.Size = 0
}

func (v *VeniceMap) Copy() *VeniceMap {
	vCopy := NewVeniceMap()
	for _, entry := range v.Entries().Values {
		entryAsTuple := entry.(*VeniceTuple)
		vCopy.Put(entryAsTuple.Values[0], entryAsTuple.Values[1])
	}
	return vCopy
}

func (v *VeniceMap) Entries() *VeniceList {
	entries := make([]VeniceValue, 0, v.Size)
	for _, chain := range v.table {
		iterator := chain
		for iterator != nil {
			entries = append(entries, &VeniceTuple{[]VeniceValue{iterator.Key, iterator.Value}})
			iterator = iterator.Next
		}
	}
	return &VeniceList{entries}
}

func (v *VeniceMap) Get(key VeniceValue) VeniceValue {
	hash := v.getHash(key)
	entry := v.table[hash]
	for entry != nil {
		if entry.Key.Equals(key) {
			return entry.Value
		}

		entry = entry.Next
	}
	return nil
}

func (v *VeniceMap) Keys() *VeniceList {
	keys := make([]VeniceValue, 0, v.Size)
	for _, chain := range v.table {
		iterator := chain
		for iterator != nil {
			keys = append(keys, iterator.Key)
			iterator = iterator.Next
		}
	}
	return &VeniceList{keys}
}

func (v *VeniceMap) Put(key VeniceValue, value VeniceValue) {
	hash := v.getHash(key)
	entry := v.table[hash]
	if entry == nil {
		v.table[hash] = &VeniceMapChain{
			Key:   key,
			Value: value,
			Next:  nil,
		}
		v.Size++
	} else {
		for entry != nil {
			if entry.Key.Equals(key) {
				entry.Value = value
				return
			}

			if entry.Next == nil {
				break
			} else {
				entry = entry.Next
			}
		}

		entry.Next = &VeniceMapChain{
			Key:   key,
			Value: value,
			Next:  nil,
		}
		v.Size++
	}
}

func (v *VeniceMap) Remove(key VeniceValue) {
	hash := v.getHash(key)
	entry := v.table[hash]
	var previous *VeniceMapChain
	for entry != nil {
		if entry.Key.Equals(key) {
			if previous == nil {
				v.table[hash] = nil
			} else {
				previous.Next = entry.Next
			}
			v.Size--
			return
		}

		previous = entry
		entry = entry.Next
	}
}

func (v *VeniceMap) Values() *VeniceList {
	values := make([]VeniceValue, 0, v.Size)
	for _, chain := range v.table {
		iterator := chain
		for iterator != nil {
			values = append(values, iterator.Value)
			iterator = iterator.Next
		}
	}
	return &VeniceList{values}
}

func (v *VeniceMap) getHash(key VeniceValue) uint64 {
	v.hash.Reset()
	// x & (y - 1) is equivalent to x % y as long as y is a power of 2.
	return key.Hash(v.hash) & uint64(len(v.table)-1)
}

func (v *VeniceBoolean) veniceValue()        {}
func (v *VeniceClassObject) veniceValue()    {}
func (v *VeniceEnumObject) veniceValue()     {}
func (v *VeniceFunctionObject) veniceValue() {}
func (v *VeniceInteger) veniceValue()        {}
func (v *VeniceList) veniceValue()           {}
func (v *VeniceListIterator) veniceValue()   {}
func (v *VeniceMap) veniceValue()            {}
func (v *VeniceMapIterator) veniceValue()    {}
func (v *VeniceRealNumber) veniceValue()     {}
func (v *VeniceString) veniceValue()         {}
func (v *VeniceTuple) veniceValue()          {}
