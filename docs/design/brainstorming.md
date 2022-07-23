This document is for brainstorming future changes to the Venice language. All ideas are preliminary and may or may not actually be implemented.

- Special annotations for functions and variables
  - e.g., for functions: pure
  - e.g., for variables: non-empty (for lists), positive (for integers)
- Exceptions or explicit error values?
  - Standard library should favor explicit error values, but exceptions should also be possible
  - Exceptions are useful while writing code because you don't need to change the return value of the function.
  - Also, exceptions print stack traces.
  - Checked exceptions?
  - <https://tour.dlang.org/tour/en/gems/uniform-function-call-syntax-ufcs>
