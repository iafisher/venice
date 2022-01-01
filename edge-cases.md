## Assigning to a parenthesized expression:
```
(a) = 42;
```

If the parser doesn't create separate nodes for parenthesized expressions, it will be impossible to disallow this.


## Assigning a variable to itself
```
const a = 42;

func main() {
  let a = a;
}
```
