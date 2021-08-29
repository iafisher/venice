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
		"int": NewConstBinding(
			&vtype.VeniceFunctionType{
				Name: "int",
				ParamTypes: []vtype.VeniceType{
					vtype.VENICE_TYPE_REAL_NUMBER,
				},
				ReturnType: vtype.VENICE_TYPE_INTEGER,
				IsBuiltin:  true,
			},
		),
		"length": NewConstBinding(
			&vtype.VeniceFunctionType{
				Name: "length",
				ParamTypes: []vtype.VeniceType{
					&vtype.VeniceUnionType{
						[]vtype.VeniceType{
							vtype.VENICE_TYPE_STRING,
							&vtype.VeniceListType{vtype.VENICE_TYPE_ANY},
							&vtype.VeniceMapType{
								KeyType:   vtype.VENICE_TYPE_ANY,
								ValueType: vtype.VENICE_TYPE_ANY,
							},
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
		"range": NewConstBinding(
			&vtype.VeniceFunctionType{
				Name: "range",
				ParamTypes: []vtype.VeniceType{
					vtype.VENICE_TYPE_INTEGER,
					vtype.VENICE_TYPE_INTEGER,
				},
				ReturnType: &vtype.VeniceListType{vtype.VENICE_TYPE_INTEGER},
				IsBuiltin:  true,
			},
		),
		"real": NewConstBinding(
			&vtype.VeniceFunctionType{
				Name: "real",
				ParamTypes: []vtype.VeniceType{
					vtype.VENICE_TYPE_INTEGER,
				},
				ReturnType: vtype.VENICE_TYPE_REAL_NUMBER,
				IsBuiltin:  true,
			},
		),
		"string": NewConstBinding(
			&vtype.VeniceFunctionType{
				Name: "string",
				ParamTypes: []vtype.VeniceType{
					&vtype.VeniceUnionType{
						[]vtype.VeniceType{
							vtype.VENICE_TYPE_BOOLEAN,
							vtype.VENICE_TYPE_CHARACTER,
							vtype.VENICE_TYPE_INTEGER,
							vtype.VENICE_TYPE_STRING,
						},
					},
				},
				ReturnType: vtype.VENICE_TYPE_STRING,
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
		"real":     NewConstBinding(vtype.VENICE_TYPE_REAL_NUMBER),
		"string":   NewConstBinding(vtype.VENICE_TYPE_STRING),
		"Optional": NewConstBinding(vtype.VENICE_TYPE_OPTIONAL),
	}
	return &SymbolTable{Parent: nil, Symbols: symbols}
}

var listBuiltins = map[string]vtype.VeniceType{
	"append": &vtype.VeniceFunctionType{
		Name:              "append",
		GenericParameters: []string{"T"},
		ParamTypes: []vtype.VeniceType{
			&vtype.VeniceListType{&vtype.VeniceSymbolType{"T"}},
			&vtype.VeniceSymbolType{"T"},
		},
		ReturnType: nil,
		IsBuiltin:  true,
	},
	"extend": &vtype.VeniceFunctionType{
		Name:              "extend",
		GenericParameters: []string{"T"},
		ParamTypes: []vtype.VeniceType{
			&vtype.VeniceListType{&vtype.VeniceSymbolType{"T"}},
			&vtype.VeniceListType{&vtype.VeniceSymbolType{"T"}},
		},
		ReturnType: nil,
		IsBuiltin:  true,
	},
	"length": &vtype.VeniceFunctionType{
		Name: "length",
		ParamTypes: []vtype.VeniceType{
			&vtype.VeniceListType{vtype.VENICE_TYPE_ANY},
		},
		ReturnType: vtype.VENICE_TYPE_INTEGER,
		IsBuiltin:  true,
	},
	"remove": &vtype.VeniceFunctionType{
		Name: "remove",
		ParamTypes: []vtype.VeniceType{
			&vtype.VeniceListType{vtype.VENICE_TYPE_ANY},
			vtype.VENICE_TYPE_INTEGER,
		},
		ReturnType: nil,
		IsBuiltin:  true,
	},
	"slice": &vtype.VeniceFunctionType{
		Name: "slice",
		ParamTypes: []vtype.VeniceType{
			&vtype.VeniceListType{&vtype.VeniceSymbolType{"T"}},
			vtype.VENICE_TYPE_INTEGER,
			vtype.VENICE_TYPE_INTEGER,
		},
		ReturnType: &vtype.VeniceListType{&vtype.VeniceSymbolType{"T"}},
		IsBuiltin:  true,
	},
}

var mapBuiltins = map[string]vtype.VeniceType{
	"entries": &vtype.VeniceFunctionType{
		Name: "entries",
		ParamTypes: []vtype.VeniceType{
			&vtype.VeniceMapType{
				KeyType:   &vtype.VeniceSymbolType{"K"},
				ValueType: &vtype.VeniceSymbolType{"V"},
			},
		},
		ReturnType: &vtype.VeniceListType{
			&vtype.VeniceTupleType{
				[]vtype.VeniceType{
					&vtype.VeniceSymbolType{"K"},
					&vtype.VeniceSymbolType{"V"},
				},
			},
		},
		IsBuiltin: true,
	},
	"keys": &vtype.VeniceFunctionType{
		Name: "keys",
		ParamTypes: []vtype.VeniceType{
			&vtype.VeniceMapType{
				KeyType:   &vtype.VeniceSymbolType{"K"},
				ValueType: vtype.VENICE_TYPE_ANY,
			},
		},
		ReturnType: &vtype.VeniceListType{&vtype.VeniceSymbolType{"K"}},
		IsBuiltin:  true,
	},
	"remove": &vtype.VeniceFunctionType{
		Name: "remove",
		ParamTypes: []vtype.VeniceType{
			&vtype.VeniceMapType{
				KeyType:   &vtype.VeniceSymbolType{"T"},
				ValueType: vtype.VENICE_TYPE_ANY,
			},
			&vtype.VeniceSymbolType{"T"},
		},
		ReturnType: nil,
		IsBuiltin:  true,
	},
	"values": &vtype.VeniceFunctionType{
		Name: "values",
		ParamTypes: []vtype.VeniceType{
			&vtype.VeniceMapType{
				KeyType:   vtype.VENICE_TYPE_ANY,
				ValueType: &vtype.VeniceSymbolType{"V"},
			},
		},
		ReturnType: &vtype.VeniceListType{&vtype.VeniceSymbolType{"V"}},
		IsBuiltin:  true,
	},
}

var stringBuiltins = map[string]vtype.VeniceType{
	"find": &vtype.VeniceFunctionType{
		Name:       "find",
		ParamTypes: []vtype.VeniceType{vtype.VENICE_TYPE_STRING, vtype.VENICE_TYPE_STRING},
		ReturnType: vtype.VeniceOptionalTypeOf(vtype.VENICE_TYPE_INTEGER),
		IsBuiltin:  true,
	},
	"find_last": &vtype.VeniceFunctionType{
		Name:       "find_last",
		ParamTypes: []vtype.VeniceType{vtype.VENICE_TYPE_STRING, vtype.VENICE_TYPE_STRING},
		ReturnType: vtype.VeniceOptionalTypeOf(vtype.VENICE_TYPE_INTEGER),
		IsBuiltin:  true,
	},
	"length": &vtype.VeniceFunctionType{
		Name:       "length",
		ParamTypes: []vtype.VeniceType{vtype.VENICE_TYPE_STRING},
		ReturnType: vtype.VENICE_TYPE_INTEGER,
		IsBuiltin:  true,
	},
	"slice": &vtype.VeniceFunctionType{
		Name: "slice",
		ParamTypes: []vtype.VeniceType{
			vtype.VENICE_TYPE_STRING,
			vtype.VENICE_TYPE_INTEGER,
			vtype.VENICE_TYPE_INTEGER,
		},
		ReturnType: vtype.VENICE_TYPE_STRING,
		IsBuiltin:  true,
	},
	"split": &vtype.VeniceFunctionType{
		Name: "split",
		ParamTypes: []vtype.VeniceType{
			vtype.VENICE_TYPE_STRING,
			vtype.VENICE_TYPE_STRING,
		},
		ReturnType: &vtype.VeniceListType{vtype.VENICE_TYPE_STRING},
		IsBuiltin:  true,
	},
	"split_space": &vtype.VeniceFunctionType{
		Name: "split_space",
		ParamTypes: []vtype.VeniceType{
			vtype.VENICE_TYPE_STRING,
		},
		ReturnType: &vtype.VeniceListType{vtype.VENICE_TYPE_STRING},
		IsBuiltin:  true,
	},
	"to_lower": &vtype.VeniceFunctionType{
		Name:       "to_lower",
		ParamTypes: []vtype.VeniceType{vtype.VENICE_TYPE_STRING},
		ReturnType: vtype.VENICE_TYPE_STRING,
		IsBuiltin:  true,
	},
	"to_upper": &vtype.VeniceFunctionType{
		Name:       "to_upper",
		ParamTypes: []vtype.VeniceType{vtype.VENICE_TYPE_STRING},
		ReturnType: vtype.VENICE_TYPE_STRING,
		IsBuiltin:  true,
	},
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
