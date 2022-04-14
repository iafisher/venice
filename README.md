# The Venice programming language
**NOTE**: Venice is in the early stages of development, and not yet ready for production use.

Venice is a modern, high-level, statically-typed programming language. It pairs the elegance and expressiveness of Python with the safety and modern language features of Rust.

<!-- Venice's syntax is closest to Rust's, so we use that as the syntax declaration for the code block. -->
```rust
import map, join from "itertools"

enum Json {
  JsonObject({string: Json}),
  JsonArray([Json]),
  JsonString(string),
  JsonNumber(real),
  JsonBoolean(bool),
  JsonNull,
}

func serialize_json(j: Json) -> string {
  match j {
    case JsonObject(obj) {
      let it = ("\(key): \(serialize_json(value))" for key, value in obj)
      return "{" ++ join(it, ", ") ++ "}"
    }
    case JsonArray(values) {
      return "[" ++ join(map(values, serialize_json), ", ") ++ "]"
    }
    case JsonString(s) {
      return s.quoted()
    }
    case JsonNumber(x) {
      return string(x)
    }
    case JsonBoolean(x) {
      return string(x)
    }
    case JsonNull {
      return "null"
    }
  }
}
```

For a full introduction to the language, read the [tutorial](https://github.com/iafisher/venice/blob/master/docs/tutorial.md).
