from attr import attrib, attrs


class VeniceAbstractType:
    pass


@attrs
class VeniceType(VeniceAbstractType):
    label = attrib()

    def __str__(self):
        return self.label


@attrs
class VeniceListType(VeniceAbstractType):
    item_type = attrib()

    def __str__(self):
        return f"list<{self.item_type}>"


@attrs
class VeniceFunctionType(VeniceAbstractType):
    parameter_types = attrib()
    return_type = attrib()
    javascript_name = attrib(default=None)

    def __str__(self):
        ptypes = ", ".join(map(str, self.parameter_types))
        return f"fn<{ptypes}, {self.return_type}>"


@attrs
class VeniceStructType(VeniceAbstractType):
    name = attrib()
    field_types = attrib()

    def __str__(self):
        return self.name


@attrs
class VeniceKeywordArgumentType(VeniceAbstractType):
    label = attrib()
    type = attrib()

    def __str__(self):
        return str(self.type)


@attrs
class VeniceMapType(VeniceAbstractType):
    key_type = attrib()
    value_type = attrib()

    def __str__(self):
        return f"map<{self.key_type}, {self.value_type}>"


VENICE_TYPE_BOOLEAN = VeniceType("boolean")
VENICE_TYPE_INTEGER = VeniceType("integer")
VENICE_TYPE_STRING = VeniceType("string")
VENICE_TYPE_VOID = VeniceType("void")
VENICE_TYPE_ANY = VeniceType("any")
