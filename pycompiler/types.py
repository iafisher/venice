from collections import namedtuple

VeniceType = namedtuple("VeniceType", ["label"])
VeniceListType = namedtuple("VeniceListType", ["item_type"])
VeniceFunctionType = namedtuple(
    "VeniceFunctionType", ["parameter_types", "return_type"]
)
VeniceStructType = namedtuple("VeniceStructType", ["name", "field_types"])
VeniceKeywordArgumentType = namedtuple("VeniceKeywordArgumentType", ["label", "type"])
VeniceMapType = namedtuple("VeniceMapType", ["key_type", "value_type"])

VENICE_TYPE_BOOLEAN = VeniceType("boolean")
VENICE_TYPE_INTEGER = VeniceType("integer")
VENICE_TYPE_STRING = VeniceType("string")
VENICE_TYPE_VOID = VeniceType("void")
VENICE_TYPE_ANY = VeniceType("any")
