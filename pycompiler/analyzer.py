from pycompiler import ast, types
from pycompiler.common import VeniceError


def vcheck(tree):
    vcheck_block(tree.statements, SymbolTable.with_globals())


def vcheck_block(statements, symbol_table, return_type=None):
    for statement in statements:
        vcheck_statement(statement, symbol_table, return_type=return_type)


def vcheck_statement(tree, symbol_table, return_type=None):
    if isinstance(tree, ast.Function):
        parameter_types = [
            types.VeniceKeywordArgumentType(p.label, resolve_type(p.type))
            for p in tree.parameters
        ]
        f_return_type = resolve_type(tree.return_type)
        symbol_table.put(
            tree.label,
            types.VeniceFunctionType(
                parameter_types=parameter_types, return_type=f_return_type,
            ),
        )

        body_symbol_table = SymbolTable(parent=symbol_table)
        for ptype in parameter_types:
            symbol_table.put(ptype.label, ptype.type)

        vcheck_block(tree.statements, body_symbol_table, return_type=f_return_type)
    elif isinstance(tree, ast.Return):
        if return_type is None:
            raise VeniceError("return statement outside of function")

        actual_return_type = vcheck_expression(tree.value, symbol_table)
        if not are_types_compatible(return_type, actual_return_type):
            raise VeniceError(
                f"expected return type of {type_to_string(return_type)}, "
                + f"got {type_to_string(actual_return_type)}"
            )
    elif isinstance(tree, ast.If):
        for clause in tree.if_clauses:
            vassert(clause.condition, symbol_table, types.VENICE_TYPE_BOOLEAN)
            vcheck_block(clause.statements, symbol_table)

        if tree.else_clause:
            vcheck_block(tree.else_clause, symbol_table)
    elif isinstance(tree, ast.Let):
        symbol_table.put(tree.label, vcheck_expression(tree.value, symbol_table))
    elif isinstance(tree, ast.Assign):
        original_type = symbol_table.get(tree.label.label)
        if original_type is None:
            raise VeniceError(f"assignment to undefined variable: {tree.label}")

        vassert(tree.value, symbol_table, original_type)
    elif isinstance(tree, ast.While):
        vassert(tree.condition, symbol_table, types.VENICE_TYPE_BOOLEAN)
        vcheck_block(tree.statements, symbol_table)
    elif isinstance(tree, ast.For):
        iterator_type = vcheck_expression(tree.iterator, symbol_table)
        if not isinstance(iterator_type, types.VeniceListType):
            raise VeniceError("loop iterator must be list")

        loop_variable_type = iterator_type.item_type
        loop_symbol_table = SymbolTable(parent=symbol_table)
        loop_symbol_table.put(tree.loop_variable, loop_variable_type)

        vcheck_block(tree.statements, loop_symbol_table)
    elif isinstance(tree, ast.ExpressionStatement):
        vcheck_expression(tree.value, symbol_table)
    elif isinstance(tree, ast.StructDeclaration):
        field_types = [
            types.VeniceKeywordArgumentType(p.label, resolve_type(p.type))
            for p in tree.fields
        ]
        symbol_table.put(
            tree.label,
            types.VeniceStructType(name=tree.label, field_types=field_types),
        )
    else:
        raise VeniceError(f"unknown AST statement type: {tree.__class__.__name__}")


def vcheck_expression(tree, symbol_table):
    if isinstance(tree, ast.Symbol):
        symbol_type = symbol_table.get(tree.label)
        if symbol_type is not None:
            return symbol_type
        else:
            raise VeniceError(f"undefined symbol: {tree.label}")
    elif isinstance(tree, ast.Infix):
        vassert(tree.left, symbol_table, types.VENICE_TYPE_INTEGER)
        vassert(tree.right, symbol_table, types.VENICE_TYPE_INTEGER)
        if tree.operator in [">=", "<=", ">", "<", "==", "!="]:
            return types.VENICE_TYPE_BOOLEAN
        else:
            return types.VENICE_TYPE_INTEGER
    elif isinstance(tree, ast.Prefix):
        if tree.operator == "not":
            vassert(tree.value, symbol_table, types.VENICE_TYPE_BOOLEAN)
            return types.VENICE_TYPE_BOOLEAN
        else:
            vassert(tree.value, symbol_table, types.VENICE_TYPE_INTEGER)
            return types.VENICE_TYPE_INTEGER
    elif isinstance(tree, ast.Call):
        function_type = vcheck_expression(tree.function, symbol_table)
        if isinstance(function_type, types.VeniceFunctionType):
            if len(function_type.parameter_types) != len(tree.arguments):
                raise VeniceError(
                    f"expected {len(function_type.parameter_types)} arguments, "
                    + f"got {len(tree.arguments)}"
                )

            for parameter, argument in zip(
                function_type.parameter_types, tree.arguments
            ):
                if isinstance(argument, ast.KeywordArgument):
                    vassert(argument.value, symbol_table, parameter.type)
                else:
                    vassert(argument, symbol_table, parameter.type)

            return function_type.return_type
        elif isinstance(function_type, types.VeniceStructType):
            for parameter, argument in zip(function_type.field_types, tree.arguments):
                if not isinstance(argument, ast.KeywordArgument):
                    raise VeniceError(
                        "struct constructor only accepts keyword arguments"
                    )

                if not argument.label == parameter.label:
                    raise VeniceError(
                        f"expected keyword argument {parameter.label}, "
                        + "got {argument.label}"
                    )

                vassert(argument.value, symbol_table, parameter.type)

            return function_type
        else:
            raise VeniceError(f"{type_to_string(function_type)} is not a function type")
    elif isinstance(tree, ast.List):
        # TODO: empty list
        item_type = vcheck_expression(tree.values[0], symbol_table)
        for value in tree.values[1:]:
            # TODO: Probably need a more robust way of checking item types, e.g.
            # collecting all types and seeing if there's a common super-type.
            another_item_type = vcheck_expression(value, symbol_table)
            if not are_types_compatible(item_type, another_item_type):
                raise VeniceError(
                    "list contains items of multiple types: "
                    + f"{type_to_string(item_type)} "
                    + f"and {type_to_string(another_item_type)}"
                )

        return types.VeniceListType(item_type)
    elif isinstance(tree, ast.Literal):
        if isinstance(tree.value, str):
            return types.VENICE_TYPE_STRING
        elif isinstance(tree.value, bool):
            # This must come before `int` because bools are ints in Python.
            return types.VENICE_TYPE_BOOLEAN
        elif isinstance(tree.value, int):
            return types.VENICE_TYPE_INTEGER
        else:
            raise VeniceError(
                f"unknown ast.Literal type: {tree.value.__class__.__name__}"
            )
    elif isinstance(tree, ast.Index):
        list_type = vcheck_expression(tree.list, symbol_table)
        index_type = vcheck_expression(tree.index, symbol_table)

        if isinstance(list_type, types.VeniceListType):
            if index_type != types.VENICE_TYPE_INTEGER:
                raise VeniceError(
                    f"index expression must be of integer type, "
                    + f"not {type_to_string(index_type)}"
                )

            return list_type.item_type
        elif isinstance(list_type, types.VeniceMapType):
            if not are_types_compatible(list_type.key_type, index_type):
                raise VeniceError(
                    f"expected {type_to_string(list_type.key_type)} for map key, "
                    + f"got {type_to_string(index_type)}"
                )

            return list_type.value_type
        else:
            raise VeniceError(f"{type_to_string(list_type)} is not a list type")
    elif isinstance(tree, ast.Map):
        key_type = vcheck_expression(tree.pairs[0].key, symbol_table)
        value_type = vcheck_expression(tree.pairs[0].value, symbol_table)
        for pair in tree.pairs[1:]:
            another_key_type = vcheck_expression(pair.key, symbol_table)
            another_value_type = vcheck_expression(pair.value, symbol_table)
            if not are_types_compatible(key_type, another_key_type):
                raise VeniceError(
                    "map contains keys of multiple types: "
                    + f"{type_to_string(key_type)} "
                    + f"and {type_to_string(another_key_type)}"
                )

            if not are_types_compatible(value_type, another_value_type):
                raise VeniceError(
                    "map contains values of multiple types: "
                    + f"{type_to_string(value_type)} "
                    + f"and {type_to_string(another_value_type)}"
                )

        return types.VeniceMapType(key_type, value_type)
    elif isinstance(tree, ast.FieldAccess):
        struct_type = vcheck_expression(tree.value, symbol_table)
        if not isinstance(struct_type, types.VeniceStructType):
            raise VeniceError(
                f"expected struct type, got {type_to_string(struct_type)}"
            )

        for field in struct_type.field_types:
            if field.label == tree.field.value:
                return field.type

        raise VeniceError(
            f"{type_to_string(struct_type)} does not have field: {tree.field.value}"
        )
    else:
        raise VeniceError(f"unknown AST expression type: {tree.__class__.__name__}")


def vassert(tree, symbol_table, expected):
    actual = vcheck_expression(tree, symbol_table)
    if not are_types_compatible(expected, actual):
        raise VeniceError(
            f"expected {type_to_string(expected)}, got {type_to_string(actual)}"
        )


def resolve_type(type_tree):
    if isinstance(type_tree, ast.Symbol):
        if type_tree.label in {"boolean", "integer", "string"}:
            return types.VeniceType(type_tree.label)
        else:
            raise VeniceError(f"unknown type: {type_tree.label}")
    elif isinstance(type_tree, ast.ParameterizedType):
        ptype = type_tree.type.value
        if ptype == "map":
            if len(type_tree.parameters) != 2:
                raise VeniceError("map type requires exactly 2 parameters")

            return types.VeniceMapType(
                resolve_type(type_tree.parameters[0]),
                resolve_type(type_tree.parameters[1]),
            )
        elif ptype == "list":
            if len(type_tree.parameters) != 1:
                raise VeniceError("list type requires exactly 1 parameter")

            return types.VeniceListType(resolve_type(type_tree.parameters[0]))
        else:
            raise VeniceError(f"{type_to_string(ptype)} cannot be parameterized")
    else:
        raise VeniceError(
            f"{type_to_string(type_tree)} cannot be interpreted as a type"
        )


def are_types_compatible(expected_type, actual_type):
    if expected_type == types.VENICE_TYPE_ANY:
        return True

    return expected_type == actual_type


def type_to_string(t):
    if isinstance(t, types.VeniceType):
        return t.label
    elif isinstance(t, types.VeniceListType):
        return f"list<{type_to_string(t.item_type)}>"
    elif isinstance(t, types.VeniceFunctionType):
        if t.parameter_types:
            return (
                "fn<"
                + ", ".join(map(type_to_string, t.parameter_types))
                + f", {type_to_string(t.return_type)}>"
            )
        else:
            return f"fn<{type_to_string(t.return_type)}>"
    elif isinstance(t, types.VeniceMapType):
        return f"map<{type_to_string(t.key_type)}, {type_to_string(t.value_type)}>"
    elif isinstance(t, types.VeniceKeywordArgumentType):
        return type_to_string(t.type)
    elif isinstance(t, types.VeniceStructType):
        return t.name
    else:
        raise ValueError(t)


class SymbolTable:
    def __init__(self, parent=None):
        self.parent = parent
        self.symbols = {}

    @classmethod
    def with_globals(cls):
        symbol_table = cls(parent=None)
        symbol_table.put(
            "print",
            types.VeniceFunctionType(
                [
                    types.VeniceKeywordArgumentType(
                        label="x", type=types.VENICE_TYPE_ANY
                    )
                ],
                return_type=types.VENICE_TYPE_VOID,
            ),
        )
        return symbol_table

    def has(self, symbol):
        if symbol in self.symbols:
            return True
        elif self.parent:
            return self.parent.has(symbol)
        else:
            return False

    def get(self, symbol):
        if symbol in self.symbols:
            return self.symbols[symbol]
        elif self.parent:
            return self.parent.get(symbol)
        else:
            return None

    def put(self, symbol, type):
        self.symbols[symbol] = type
