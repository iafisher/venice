package main

type VeniceType interface {
	veniceType()
}

type VeniceFunctionType struct {
	ParamTypes []VeniceType
	ReturnType VeniceType
}

func (t *VeniceFunctionType) veniceType() {}

type VeniceMapType struct {
	KeyType   VeniceType
	ValueType VeniceType
}

func (t *VeniceMapType) veniceType() {}

type VeniceListType struct {
	ItemType VeniceType
}

func (t *VeniceListType) veniceType() {}

type VeniceAtomicType struct {
	Type string
}

func (t *VeniceAtomicType) veniceType() {}
