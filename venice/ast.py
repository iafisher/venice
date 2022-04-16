from abc import ABC
from dataclasses import dataclass
from typing import List, Optional

##
# Base classes
##


class Node(ABC):
    def accept(self, visitor):
        visit_function_name = f"visit_{self.__class__.__name__}"
        visit_function = getattr(visitor, visit_function_name)
        return visit_function(self)


class Expression(Node):
    pass


class Statement(Node):
    pass


class Type(Node):
    pass


##
# Top-level nodes
##


@dataclass
class Module(Node):
    functions: List["Function"]


@dataclass
class Function(Node):
    name: str
    parameters: List["Parameter"]
    body: List[Statement]
    return_type: Optional[Type] = None


@dataclass
class Parameter(Node):
    name: str
    type: Type


##
# Statements
##


@dataclass
class ExpressionStatement(Statement):
    expression: Expression


@dataclass
class Let(Statement):
    symbol: str
    value: Expression


@dataclass
class Return(Statement):
    value: Optional[Expression]


##
# Expressions
##


@dataclass
class FunctionCall(Expression):
    function: str
    arguments: List[Expression]


@dataclass
class Infix(Expression):
    operator: str
    left: Expression
    right: Expression


@dataclass
class Integer(Expression):
    value: int


@dataclass
class String(Expression):
    value: str


@dataclass
class Symbol(Expression):
    value: str


##
# Types
##


@dataclass
class SymbolType(Type):
    value: str
