from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Dict, List, Optional

from . import ast, common


def codegen(tree: ast.Node) -> str:
    generator = CodeGenerator()
    return generator.generate(tree)


class VeniceType(ABC):
    @abstractmethod
    def match(self, other: "VeniceType") -> bool:
        pass

    @abstractmethod
    def c_type(self) -> str:
        pass


class SymbolTable:
    symbols: Dict[str, VeniceType]
    enclosing: Optional["SymbolTable"]

    def __init__(self, enclosing: Optional["SymbolTable"] = None) -> None:
        self.symbols = {}
        self.enclosing = enclosing

    def get(self, name: str) -> VeniceType:
        r = self.symbols.get(name)
        if r is None:
            if self.enclosing:
                return self.enclosing.get(name)

            raise common.VeniceTypeError(f"undefined symbol: {name}")
        return r

    def put(self, name: str, type_: VeniceType) -> None:
        if name in self.symbols:
            raise common.VeniceTypeError(f"redefinition of symbol: {name}")

        self.symbols[name] = type_


class CodeGenerator:
    function_typedefs: List["VeniceFunctionType"]
    symbol_table: SymbolTable

    def __init__(self) -> None:
        self.function_typedefs = []
        self.symbol_table = SymbolTable()

    def generate(self, node: ast.Node) -> str:
        return node.accept(self)

    def visit_Module(self, node: ast.Module) -> str:
        builder = []
        for function in node.functions:
            builder.append(self.generate(function))

        return (
            "#include <venice.h>\n\n"
            + "\n".join(t.to_typedef() for t in self.function_typedefs)
            + "\n\n"
            + "\n\n".join(builder)
        )

    def visit_Function(self, node: ast.Function) -> str:
        if node.return_type is None:
            return_type = VENICE_VOID_TYPE
        else:
            return_type = self.resolve_ast_type(node.return_type)

        builder = []
        builder.append(return_type.c_type())
        builder.append(" ")
        builder.append(node.name)
        builder.append("(")

        function_symbol_table = SymbolTable(enclosing=self.symbol_table)
        parameter_types = []
        for i, param in enumerate(node.parameters):
            param_type = self.resolve_ast_type(param.type)
            parameter_types.append(param_type)
            function_symbol_table.put(param.name, param_type)

            builder.append(param_type.c_type())
            builder.append(" ")
            builder.append(param.name)

            if i != len(node.parameters) - 1:
                builder.append(", ")

        function_type = VeniceFunctionType(
            parameter_types=parameter_types,
            return_type=return_type,
        )
        self.function_typedefs.append(function_type)
        # Put the function's name in the symbol table before generating code for the
        # body, so that recursive functions can call themselves.
        self.symbol_table.put(node.name, function_type)

        builder.append(") {\n")
        self.symbol_table = function_symbol_table
        for statement in node.body:
            code = self.generate(statement)
            builder.append("  " + code + "\n")
        self.symbol_table = self.symbol_table.enclosing
        builder.append("}\n")
        return "".join(builder)

    def visit_Let(self, node: ast.Let) -> str:
        value_type = self.compute_type(node.value)
        self.symbol_table.put(node.symbol, value_type)

        value_code = self.generate(node.value)
        return f"{value_type.c_type()} {node.symbol} = {value_code};"

    def visit_Infix(self, node: ast.Infix) -> str:
        return f"{self.generate(node.left)} {node.operator} {self.generate(node.right)}"

    def visit_Integer(self, node: ast.Integer) -> str:
        return str(node.value)

    def visit_Symbol(self, node: ast.Symbol) -> str:
        return node.value

    def visit_String(self, node: ast.String) -> str:
        return repr(node.value)

    def compute_type(self, node: ast.Expression) -> VeniceType:
        if isinstance(node, ast.Infix):
            self.assert_type(node.left, VENICE_INT_TYPE)
            self.assert_type(node.right, VENICE_INT_TYPE)
            return VENICE_INT_TYPE
        elif isinstance(node, ast.Integer):
            return VENICE_INT_TYPE
        elif isinstance(node, ast.String):
            return VENICE_STRING_TYPE
        elif isinstance(node, ast.Symbol):
            return self.symbol_table.get(node.value)
        else:
            raise common.VeniceTypeError(f"unknown expression type: {node!r}")

    def assert_type(self, node: ast.Expression, expected: VeniceType) -> VeniceType:
        actual = self.compute_type(node)
        if not expected.match(actual):
            raise common.VeniceTypeError(
                f"expected {expected!r}, got {actual!r} for {node!r}"
            )
        return actual

    def resolve_ast_type(self, node: ast.Type) -> VeniceType:
        if isinstance(node, ast.SymbolType):
            # TODO: Have a proper symbol table for type look-up.
            if node.value == "bool":
                return VENICE_BOOL_TYPE
            elif node.value == "int":
                return VENICE_INT_TYPE
            elif node.value == "str":
                return VENICE_STRING_TYPE
            else:
                raise common.VeniceTypeError(f"unknown type name: {node.value!r}")
        else:
            raise common.VeniceTypeError(f"unknown type node: {node!r}")


@dataclass
class VeniceFunctionType(VeniceType):
    parameter_types: List[VeniceType]
    return_type: VeniceType

    def match(self, other: VeniceType) -> bool:
        if not isinstance(other, VeniceFunctionType):
            return False

        if len(self.parameter_types) != len(other.parameter_types):
            return False

        if not self.return_type.match(other.return_type):
            return False

        for my_type, other_type in zip(self.parameter_types, other.parameter_types):
            if not my_type.match(other_type):
                return False

        return True

    def c_type(self) -> str:
        # This type name is created with a typedef by CodeGenerator.
        return "__".join(
            [t.c_type() for t in self.parameter_types] + [self.return_type.c_type()]
        )

    def to_typedef(self) -> str:
        typedef_builder = []
        typedef_builder.append("typedef ")
        typedef_builder.append(self.return_type.c_type())
        typedef_builder.append(" ")
        typedef_builder.append(self.c_type())
        typedef_builder.append("(")
        for i, param_type in enumerate(self.parameter_types):
            typedef_builder.append(param_type.c_type())
            if i != len(self.parameter_types) - 1:
                typedef_builder.append(", ")
        typedef_builder.append(");")
        return "".join(typedef_builder)


@dataclass
class VeniceAtomicType(VeniceType):
    name: str

    def match(self, other: VeniceType) -> bool:
        return isinstance(other, VeniceAtomicType) and self.name == other.name

    def c_type(self) -> str:
        return self.name


VENICE_BOOL_TYPE = VeniceAtomicType("venice_bool_t")
VENICE_INT_TYPE = VeniceAtomicType("venice_int_t")
VENICE_VOID_TYPE = VeniceAtomicType("void")
VENICE_STRING_TYPE = VeniceAtomicType("venice_string_t")
