package vm

import (
	"testing"
)

func TestHashMapIntToInt(t *testing.T) {
	hashMap := NewVeniceMap()

	for i := 0; i < 1000; i++ {
		hashMap.Put(I(i), I(i+1))
	}

	if hashMap.Size != 1000 {
		t.Fatalf("Expected size of 1000, got %q", hashMap.Size)
	}

	for key := 999; key >= 0; key-- {
		value := hashMap.Get(I(key))
		if !value.Equals(I(key + 1)) {
			t.Fatalf("Expected %q, got %s", key+1, value.String())
		}
	}

	hashMap.Put(I(100), I(0))
	value := hashMap.Get(I(100))
	if !value.Equals(I(0)) {
		t.Fatalf("Expected 0, got %s", value.String())
	}

	if hashMap.Size != 1000 {
		t.Fatalf("Expected size of 1000, got %q", hashMap.Size)
	}

	hashMap.Remove(I(24))
	hashMap.Remove(I(48))
	hashMap.Remove(I(96))

	if hashMap.Size != 997 {
		t.Fatalf("Expected size of 997, got %q", hashMap.Size)
	}

	value = hashMap.Get(I(24))
	if value != nil {
		t.Fatalf("Expected nil, got %s", value.String())
	}
}

func TestHashMapStringToString(t *testing.T) {
	hashMap := NewVeniceMap()

	hashMap.Put(S("Canada"), S("Ottawa"))
	hashMap.Put(S("United States"), S("Washington"))
	hashMap.Put(S("Mexico"), S("Mexico City"))
	hashMap.Put(S("Cuba"), S("Havana"))
	hashMap.Put(S("Guatemala"), S("Guatemala City"))

	if hashMap.Size != 5 {
		t.Fatalf("Expected size of 5, got %q", hashMap.Size)
	}

	value := hashMap.Get(S("Cuba"))
	if !value.Equals(S("Havana")) {
		t.Fatalf("Expected \"Havana\", got %s", value.String())
	}

	value = hashMap.Get(S("Nicaragua"))
	if value != nil {
		t.Fatalf("Expected nil, got %s", value.String())
	}

	hashMap.Remove(S("Guatemala"))
	if hashMap.Size != 4 {
		t.Fatalf("Expected size of 4, got %q", hashMap.Size)
	}
}

func I(x int) *VeniceInteger {
	return &VeniceInteger{x}
}

func S(s string) *VeniceString {
	return &VeniceString{s}
}
