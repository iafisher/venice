from dataclasses import dataclass


@dataclass
class Location:
    file: str
    line: int
    column: int

    def __str__(self) -> str:
        return f"line {self.line}, column {self.column} of {self.file}"


class VeniceError(Exception):
    pass


class VeniceInternalError(Exception):
    pass


class VeniceSyntaxError(VeniceError):
    pass


class VeniceTypeError(VeniceError):
    pass
