from attr import attrib, attrs


class AbstractNode:
    pass


@attrs
class ProgramNode(AbstractNode):
    statements = attrib()


@attrs
class FunctionNode(AbstractNode):
    label = attrib()
    parameters = attrib()
    return_type = attrib()
    statements = attrib()


@attrs
class ParameterNode(AbstractNode):
    label = attrib()
    type = attrib()


@attrs
class LetNode(AbstractNode):
    label = attrib()
    value = attrib()


@attrs
class ReturnNode(AbstractNode):
    value = attrib()


@attrs
class IfNode(AbstractNode):
    if_clauses = attrib()
    else_clause = attrib()


@attrs
class IfClauseNode(AbstractNode):
    condition = attrib()
    statements = attrib()


@attrs
class WhileNode(AbstractNode):
    condition = attrib()
    statements = attrib()


@attrs
class ForNode(AbstractNode):
    loop_variable = attrib()
    iterator = attrib()
    statements = attrib()


@attrs
class AssignNode(AbstractNode):
    label = attrib()
    value = attrib()


@attrs
class StructDeclarationNode(AbstractNode):
    label = attrib()
    fields = attrib()


@attrs
class StructDeclarationFieldNode(AbstractNode):
    label = attrib()
    type = attrib()


@attrs
class ExpressionStatementNode(AbstractNode):
    value = attrib()


@attrs
class CallNode(AbstractNode):
    function = attrib()
    arguments = attrib()


@attrs
class IndexNode(AbstractNode):
    list = attrib()
    index = attrib()


@attrs
class KeywordArgumentNode(AbstractNode):
    label = attrib()
    value = attrib()


@attrs
class InfixNode(AbstractNode):
    operator = attrib()
    left = attrib()
    right = attrib()


@attrs
class PrefixNode(AbstractNode):
    operator = attrib()
    value = attrib()


@attrs
class SymbolNode(AbstractNode):
    label = attrib()


@attrs
class LiteralNode(AbstractNode):
    value = attrib()


@attrs
class ListNode(AbstractNode):
    values = attrib()


@attrs
class MapNode(AbstractNode):
    pairs = attrib()


@attrs
class MapLiteralPairNode(AbstractNode):
    key = attrib()
    value = attrib()


@attrs
class ParameterizedTypeNode(AbstractNode):
    type = attrib()
    parameters = attrib()


@attrs
class FieldAccessNode(AbstractNode):
    value = attrib()
    field = attrib()
