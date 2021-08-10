package main

type VeniceType interface {
	veniceType()
}

type VeniceClassType struct {
	Fields []*VeniceClassField
}

func (t *VeniceClassType) veniceType() {}

type VeniceClassField struct {
	Name      string
	Public    bool
	FieldType VeniceType
}

type VeniceEnumType struct {
	Cases []*VeniceCaseType
}

func (t *VeniceEnumType) veniceType() {}

type VeniceCaseType struct {
	Label string
	Types []VeniceType
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
