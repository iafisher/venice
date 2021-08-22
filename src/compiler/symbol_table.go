package compiler

import (
	"github.com/iafisher/venice/src/vtype"
)

type SymbolTable struct {
	Parent  *SymbolTable
	Symbols map[string]*SymbolTableBinding
}

type SymbolTableBinding struct {
	IsVar bool
	Type  vtype.VeniceType
}

func NewConstBinding(sType vtype.VeniceType) *SymbolTableBinding {
	return &SymbolTableBinding{IsVar: false, Type: sType}
}

func NewVarBinding(sType vtype.VeniceType) *SymbolTableBinding {
	return &SymbolTableBinding{IsVar: true, Type: sType}
}

func NewBuiltinSymbolTable() *SymbolTable {
	symbols := map[string]*SymbolTableBinding{
		"length": NewConstBinding(
			&vtype.VeniceFunctionType{
				Name: "length",
				ParamTypes: []vtype.VeniceType{
					&vtype.VeniceUnionType{
						[]vtype.VeniceType{
							vtype.VENICE_TYPE_STRING,
							&vtype.VeniceListType{vtype.VENICE_TYPE_ANY},
							&vtype.VeniceMapType{vtype.VENICE_TYPE_ANY, vtype.VENICE_TYPE_ANY},
						},
					},
				},
				ReturnType: vtype.VENICE_TYPE_INTEGER,
				IsBuiltin:  true,
			},
		),
		"print": NewConstBinding(
			&vtype.VeniceFunctionType{
				Name:       "print",
				ParamTypes: []vtype.VeniceType{vtype.VENICE_TYPE_ANY},
				ReturnType: nil,
				IsBuiltin:  true,
			},
		),
	}
	return &SymbolTable{Parent: nil, Symbols: symbols}
}

func NewBuiltinTypeSymbolTable() *SymbolTable {
	symbols := map[string]*SymbolTableBinding{
		"bool":     NewConstBinding(vtype.VENICE_TYPE_BOOLEAN),
		"int":      NewConstBinding(vtype.VENICE_TYPE_INTEGER),
		"string":   NewConstBinding(vtype.VENICE_TYPE_STRING),
		"Optional": NewConstBinding(vtype.VENICE_TYPE_OPTIONAL),
	}
	return &SymbolTable{Parent: nil, Symbols: symbols}
}

func (symtab *SymbolTable) SpawnChild() *SymbolTable {
	return &SymbolTable{
		Parent:  symtab,
		Symbols: map[string]*SymbolTableBinding{},
	}
}

func (symtab *SymbolTable) GetBinding(symbol string) (*SymbolTableBinding, bool) {
	binding, ok := symtab.Symbols[symbol]
	if !ok {
		if symtab.Parent != nil {
			return symtab.Parent.GetBinding(symbol)
		} else {
			return nil, false
		}
	}
	return binding, true
}

func (symtab *SymbolTable) Get(symbol string) (vtype.VeniceType, bool) {
	binding, ok := symtab.Symbols[symbol]
	if !ok {
		if symtab.Parent != nil {
			return symtab.Parent.Get(symbol)
		} else {
			return nil, false
		}
	}
	return binding.Type, true
}

func (symtab *SymbolTable) Put(symbol string, value vtype.VeniceType) {
	symtab.Symbols[symbol] = NewConstBinding(value)
}

func (symtab *SymbolTable) PutVar(symbol string, value vtype.VeniceType) {
	symtab.Symbols[symbol] = NewVarBinding(value)
}