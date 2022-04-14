from dataclasses import dataclass


@dataclass
class Location:
    file: str
    line: int
    column: int


class VeniceError(Exception):
    pass


class VeniceParseError(VeniceError):
    pass


class VeniceTypeError(VeniceError):
    pass
