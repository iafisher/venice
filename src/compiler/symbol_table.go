package compiler

type SymbolTable struct {
	Parent  *SymbolTable
	Symbols map[string]*SymbolTableBinding
}

type SymbolTableBinding struct {
	IsVar bool
	Type  VeniceType
}

func NewConstBinding(sType VeniceType) *SymbolTableBinding {
	return &SymbolTableBinding{IsVar: false, Type: sType}
}

func NewVarBinding(sType VeniceType) *SymbolTableBinding {
	return &SymbolTableBinding{IsVar: true, Type: sType}
}

func NewBuiltinSymbolTable() *SymbolTable {
	symbols := map[string]*SymbolTableBinding{
		"input": NewConstBinding(
			&VeniceFunctionType{
				Name: "input",
				ParamTypes: []VeniceType{
					VENICE_TYPE_STRING,
				},
				ReturnType: VeniceOptionalTypeOf(VENICE_TYPE_STRING),
				IsBuiltin:  true,
			},
		),
		"int": NewConstBinding(
			&VeniceFunctionType{
				Name: "int",
				ParamTypes: []VeniceType{
					VENICE_TYPE_REAL_NUMBER,
				},
				ReturnType: VENICE_TYPE_INTEGER,
				IsBuiltin:  true,
			},
		),
		"length": NewConstBinding(
			&VeniceFunctionType{
				Name: "length",
				ParamTypes: []VeniceType{
					&VeniceUnionType{
						[]VeniceType{
							VENICE_TYPE_STRING,
							&VeniceListType{VENICE_TYPE_ANY},
							&VeniceMapType{
								KeyType:   VENICE_TYPE_ANY,
								ValueType: VENICE_TYPE_ANY,
							},
						},
					},
				},
				ReturnType: VENICE_TYPE_INTEGER,
				IsBuiltin:  true,
			},
		),
		"print": NewConstBinding(
			&VeniceFunctionType{
				Name:       "print",
				ParamTypes: []VeniceType{VENICE_TYPE_ANY},
				ReturnType: nil,
				IsBuiltin:  true,
			},
		),
		"range": NewConstBinding(
			&VeniceFunctionType{
				Name: "range",
				ParamTypes: []VeniceType{
					VENICE_TYPE_INTEGER,
					VENICE_TYPE_INTEGER,
				},
				ReturnType: &VeniceListType{VENICE_TYPE_INTEGER},
				IsBuiltin:  true,
			},
		),
		"real": NewConstBinding(
			&VeniceFunctionType{
				Name: "real",
				ParamTypes: []VeniceType{
					VENICE_TYPE_INTEGER,
				},
				ReturnType: VENICE_TYPE_REAL_NUMBER,
				IsBuiltin:  true,
			},
		),
		"string": NewConstBinding(
			&VeniceFunctionType{
				Name: "string",
				ParamTypes: []VeniceType{
					&VeniceUnionType{
						[]VeniceType{
							VENICE_TYPE_BOOLEAN,
							VENICE_TYPE_CHARACTER,
							VENICE_TYPE_INTEGER,
							VENICE_TYPE_STRING,
						},
					},
				},
				ReturnType: VENICE_TYPE_STRING,
				IsBuiltin:  true,
			},
		),
	}
	return &SymbolTable{Parent: nil, Symbols: symbols}
}

func NewBuiltinTypeSymbolTable() *SymbolTable {
	symbols := map[string]*SymbolTableBinding{
		"bool":     NewConstBinding(VENICE_TYPE_BOOLEAN),
		"int":      NewConstBinding(VENICE_TYPE_INTEGER),
		"real":     NewConstBinding(VENICE_TYPE_REAL_NUMBER),
		"string":   NewConstBinding(VENICE_TYPE_STRING),
		"Optional": NewConstBinding(VENICE_TYPE_OPTIONAL),
	}
	return &SymbolTable{Parent: nil, Symbols: symbols}
}

var listBuiltins = map[string]VeniceType{
	"append": &VeniceFunctionType{
		Name:              "append",
		GenericParameters: []string{"T"},
		ParamTypes: []VeniceType{
			&VeniceListType{&VeniceSymbolType{"T"}},
			&VeniceSymbolType{"T"},
		},
		ReturnType: nil,
		IsBuiltin:  true,
	},
	"copy": &VeniceFunctionType{
		Name: "copy",
		ParamTypes: []VeniceType{
			&VeniceListType{&VeniceSymbolType{"T"}},
		},
		ReturnType: &VeniceListType{&VeniceSymbolType{"T"}},
		IsBuiltin:  true,
	},
	"extend": &VeniceFunctionType{
		Name:              "extend",
		GenericParameters: []string{"T"},
		ParamTypes: []VeniceType{
			&VeniceListType{&VeniceSymbolType{"T"}},
			&VeniceListType{&VeniceSymbolType{"T"}},
		},
		ReturnType: nil,
		IsBuiltin:  true,
	},
	"find": &VeniceFunctionType{
		Name: "find",
		ParamTypes: []VeniceType{
			&VeniceListType{&VeniceSymbolType{"T"}},
			&VeniceSymbolType{"T"},
		},
		ReturnType: VeniceOptionalTypeOf(VENICE_TYPE_INTEGER),
		IsBuiltin:  true,
	},
	"find_last": &VeniceFunctionType{
		Name: "find_last",
		ParamTypes: []VeniceType{
			&VeniceListType{&VeniceSymbolType{"T"}},
			&VeniceSymbolType{"T"},
		},
		ReturnType: VeniceOptionalTypeOf(VENICE_TYPE_INTEGER),
		IsBuiltin:  true,
	},
	"join": &VeniceFunctionType{
		Name: "join",
		ParamTypes: []VeniceType{
			&VeniceListType{VENICE_TYPE_STRING},
			VENICE_TYPE_STRING,
		},
		ReturnType: VENICE_TYPE_STRING,
		IsBuiltin:  true,
	},
	"length": &VeniceFunctionType{
		Name: "length",
		ParamTypes: []VeniceType{
			&VeniceListType{&VeniceSymbolType{"T"}},
		},
		ReturnType: VENICE_TYPE_INTEGER,
		IsBuiltin:  true,
	},
	"remove": &VeniceFunctionType{
		Name: "remove",
		ParamTypes: []VeniceType{
			&VeniceListType{&VeniceSymbolType{"T"}},
			VENICE_TYPE_INTEGER,
		},
		ReturnType: nil,
		IsBuiltin:  true,
	},
	"reversed": &VeniceFunctionType{
		Name: "reversed",
		ParamTypes: []VeniceType{
			&VeniceListType{&VeniceSymbolType{"T"}},
		},
		ReturnType: &VeniceListType{&VeniceSymbolType{"T"}},
		IsBuiltin:  true,
	},
	"reverse_in_place": &VeniceFunctionType{
		Name: "reverse_in_place",
		ParamTypes: []VeniceType{
			&VeniceListType{&VeniceSymbolType{"T"}},
		},
		ReturnType: nil,
		IsBuiltin:  true,
	},
	"slice": &VeniceFunctionType{
		Name: "slice",
		ParamTypes: []VeniceType{
			&VeniceListType{&VeniceSymbolType{"T"}},
			VENICE_TYPE_INTEGER,
			VENICE_TYPE_INTEGER,
		},
		ReturnType: &VeniceListType{&VeniceSymbolType{"T"}},
		IsBuiltin:  true,
	},
	"sorted": &VeniceFunctionType{
		Name: "sorted",
		ParamTypes: []VeniceType{
			&VeniceListType{&VeniceSymbolType{"T"}},
		},
		ReturnType: &VeniceListType{&VeniceSymbolType{"T"}},
		IsBuiltin:  true,
	},
	"sort_in_place": &VeniceFunctionType{
		Name: "sort_in_place",
		ParamTypes: []VeniceType{
			&VeniceListType{&VeniceSymbolType{"T"}},
		},
		ReturnType: nil,
		IsBuiltin:  true,
	},
}

var mapBuiltins = map[string]VeniceType{
	"clear": &VeniceFunctionType{
		Name: "clear",
		ParamTypes: []VeniceType{
			&VeniceMapType{
				KeyType:   &VeniceSymbolType{"K"},
				ValueType: &VeniceSymbolType{"V"},
			},
		},
		ReturnType: nil,
		IsBuiltin:  true,
	},
	"copy": &VeniceFunctionType{
		Name: "copy",
		ParamTypes: []VeniceType{
			&VeniceMapType{
				KeyType:   &VeniceSymbolType{"K"},
				ValueType: &VeniceSymbolType{"V"},
			},
		},
		ReturnType: &VeniceMapType{
			KeyType:   &VeniceSymbolType{"K"},
			ValueType: &VeniceSymbolType{"V"},
		},
		IsBuiltin: true,
	},
	"entries": &VeniceFunctionType{
		Name: "entries",
		ParamTypes: []VeniceType{
			&VeniceMapType{
				KeyType:   &VeniceSymbolType{"K"},
				ValueType: &VeniceSymbolType{"V"},
			},
		},
		ReturnType: &VeniceListType{
			&VeniceTupleType{
				[]VeniceType{
					&VeniceSymbolType{"K"},
					&VeniceSymbolType{"V"},
				},
			},
		},
		IsBuiltin: true,
	},
	"keys": &VeniceFunctionType{
		Name: "keys",
		ParamTypes: []VeniceType{
			&VeniceMapType{
				KeyType:   &VeniceSymbolType{"K"},
				ValueType: VENICE_TYPE_ANY,
			},
		},
		ReturnType: &VeniceListType{&VeniceSymbolType{"K"}},
		IsBuiltin:  true,
	},
	"remove": &VeniceFunctionType{
		Name: "remove",
		ParamTypes: []VeniceType{
			&VeniceMapType{
				KeyType:   &VeniceSymbolType{"K"},
				ValueType: &VeniceSymbolType{"V"},
			},
			&VeniceSymbolType{"K"},
		},
		ReturnType: nil,
		IsBuiltin:  true,
	},
	"values": &VeniceFunctionType{
		Name: "values",
		ParamTypes: []VeniceType{
			&VeniceMapType{
				KeyType:   &VeniceSymbolType{"K"},
				ValueType: &VeniceSymbolType{"V"},
			},
		},
		ReturnType: &VeniceListType{&VeniceSymbolType{"V"}},
		IsBuiltin:  true,
	},
}

var stringBuiltins = map[string]VeniceType{
	"find": &VeniceFunctionType{
		Name:       "find",
		ParamTypes: []VeniceType{VENICE_TYPE_STRING, VENICE_TYPE_STRING},
		ReturnType: VeniceOptionalTypeOf(VENICE_TYPE_INTEGER),
		IsBuiltin:  true,
	},
	"find_last": &VeniceFunctionType{
		Name:       "find_last",
		ParamTypes: []VeniceType{VENICE_TYPE_STRING, VENICE_TYPE_STRING},
		ReturnType: VeniceOptionalTypeOf(VENICE_TYPE_INTEGER),
		IsBuiltin:  true,
	},
	"length": &VeniceFunctionType{
		Name:       "length",
		ParamTypes: []VeniceType{VENICE_TYPE_STRING},
		ReturnType: VENICE_TYPE_INTEGER,
		IsBuiltin:  true,
	},
	"quoted": &VeniceFunctionType{
		Name:       "quoted",
		ParamTypes: []VeniceType{VENICE_TYPE_STRING},
		ReturnType: VENICE_TYPE_STRING,
		IsBuiltin:  true,
	},
	"slice": &VeniceFunctionType{
		Name: "slice",
		ParamTypes: []VeniceType{
			VENICE_TYPE_STRING,
			VENICE_TYPE_INTEGER,
			VENICE_TYPE_INTEGER,
		},
		ReturnType: VENICE_TYPE_STRING,
		IsBuiltin:  true,
	},
	"split": &VeniceFunctionType{
		Name: "split",
		ParamTypes: []VeniceType{
			VENICE_TYPE_STRING,
			VENICE_TYPE_STRING,
		},
		ReturnType: &VeniceListType{VENICE_TYPE_STRING},
		IsBuiltin:  true,
	},
	"split_space": &VeniceFunctionType{
		Name: "split_space",
		ParamTypes: []VeniceType{
			VENICE_TYPE_STRING,
		},
		ReturnType: &VeniceListType{VENICE_TYPE_STRING},
		IsBuiltin:  true,
	},
	"to_lower": &VeniceFunctionType{
		Name:       "to_lower",
		ParamTypes: []VeniceType{VENICE_TYPE_STRING},
		ReturnType: VENICE_TYPE_STRING,
		IsBuiltin:  true,
	},
	"to_upper": &VeniceFunctionType{
		Name:       "to_upper",
		ParamTypes: []VeniceType{VENICE_TYPE_STRING},
		ReturnType: VENICE_TYPE_STRING,
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

func (symtab *SymbolTable) Get(symbol string) (VeniceType, bool) {
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

func (symtab *SymbolTable) Put(symbol string, value VeniceType) {
	symtab.Symbols[symbol] = NewConstBinding(value)
}

func (symtab *SymbolTable) PutVar(symbol string, value VeniceType) {
	symtab.Symbols[symbol] = NewVarBinding(value)
}
