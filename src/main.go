package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		repl_vm()
	} else if len(os.Args) == 2 {
		switch os.Args[1] {
		case "lex":
			repl_lexer()
		case "parse":
			repl_parser()
		case "compile":
			repl_compiler()
		case "execute":
			repl_vm()
		default:
			fmt.Printf("Error: unknown subcommand %q", os.Args[1])
		}
	} else {
		fmt.Println("Error: too many command-line arguments.")
	}
}

func repl_lexer() {
	repl_generic(func(line string) {
		lexer := NewLexer(line)
		token := lexer.NextToken()
		for token.Type != TOKEN_EOF {
			fmt.Println(token.asString())
			token = lexer.NextToken()
		}
	})
}

func repl_parser() {
	repl_generic(func(line string) {
		tree, ok := NewParser(NewLexer(line)).Parse()
		if ok {
			fmt.Printf("%+v\n", tree)
		} else {
			fmt.Println("Parse error")
		}
	})
}

func repl_compiler() {
	compiler := NewCompiler()
	repl_generic(func(line string) {
		tree, ok := NewParser(NewLexer(line)).Parse()
		if !ok {
			fmt.Println("Parse error")
		}

		bytecodes, ok := compiler.Compile(tree)
		if !ok {
			fmt.Println("Compile error")
		}

		for _, bytecode := range bytecodes {
			fmt.Print(bytecode.Name)
			for _, arg := range bytecode.Args {
				fmt.Printf(" %+v", arg)
			}
			fmt.Print("\n")
		}
	})
}

func repl_vm() {
	vm := NewVirtualMachine()
	compiler := NewCompiler()
	repl_generic(func(line string) {
		tree, ok := NewParser(NewLexer(line)).Parse()
		if !ok {
			fmt.Println("Parse error")
		}

		bytecodes, ok := compiler.Compile(tree)
		if !ok {
			fmt.Println("Compile error")
		}

		value, ok := vm.Execute(bytecodes)
		if !ok {
			fmt.Println("Execution error")
		}

		if value != nil {
			switch typedValue := value.(type) {
			case *VeniceInteger:
				fmt.Printf("%d\n", typedValue.Value)
			default:
				fmt.Printf("%+v\n", value)
			}
		}
	})
}

func repl_generic(action func(line string)) {
	fmt.Print("The Venice programming language.\n\n")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(">>> ")
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		action(line)
	}
}
