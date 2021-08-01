package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
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
	} else if len(os.Args) == 3 {
		switch os.Args[1] {
		case "compile":
			compile_program(os.Args[2])
		case "execute":
			execute_program(os.Args[2])
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
		tree, err := NewParser(NewLexer(line)).Parse()
		if err == nil {
			fmt.Printf("%+v\n", tree)
		} else {
			fmt.Printf("Parse error: %v\n", err)
		}
	})
}

func repl_compiler() {
	compiler := NewCompiler()
	repl_generic(func(line string) {
		tree, err := NewParser(NewLexer(line)).Parse()
		if err != nil {
			fmt.Printf("Parse error: %v\n", err)
			return
		}

		bytecodes, err := compiler.Compile(tree)
		if err != nil {
			fmt.Printf("Compile error: %v\n", err)
			return
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
		tree, err := NewParser(NewLexer(line)).Parse()
		if err != nil {
			fmt.Printf("Parse error: %v\n", err)
			return
		}

		bytecodes, err := compiler.Compile(tree)
		if err != nil {
			fmt.Printf("Compile error: %v\n", err)
			return
		}

		value, err := vm.Execute(bytecodes)
		if err != nil {
			fmt.Printf("Execution error: %v\n", err)
			return
		}

		if value != nil {
			fmt.Println(value.Serialize())
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

func compile_program(p string) {
	data, err := ioutil.ReadFile(p)
	if err != nil {
		log.Fatal(err)
	}

	program := string(data)
	tree, err := NewParser(NewLexer(program)).Parse()
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	bytecodes, err := NewCompiler().Compile(tree)
	if err != nil {
		log.Fatalf("Compile error: %v", err)
	}

	ext := path.Ext(p)
	outputPath := p[:len(p)-len(ext)] + ".vnb"

	f, err := os.Create(outputPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	writer := bufio.NewWriter(f)

	for _, bytecode := range bytecodes {
		writer.WriteString(bytecode.Name)
		for _, arg := range bytecode.Args {
			writer.WriteString(" ")
			writer.WriteString(arg.Serialize())
		}
		writer.WriteString("\n")
	}

	writer.Flush()
}

func execute_program(p string) {
	if strings.HasSuffix(p, ".vn") {
		log.Fatal("Error: can only execute compiled programs.")
	}

	data, err := ioutil.ReadFile(p)
	if err != nil {
		log.Fatal(err)
	}

	program := string(data)
	bytecodes := []*Bytecode{}
	for i, line := range strings.Split(program, "\n") {
		lexer := NewLexer(line)
		instruction := lexer.NextToken()
		if instruction.Type == TOKEN_EOF {
			continue
		}

		if instruction.Type != TOKEN_SYMBOL {
			log.Fatalf("Could not parse line %d", i+1)
		}

		args := []VeniceValue{}
		token := lexer.NextToken()
		for token.Type != TOKEN_EOF {
			switch token.Type {
			case TOKEN_INT:
				value, err := strconv.ParseInt(token.Value, 10, 64)
				if err != nil {
					log.Fatal("Could not parse integer token")
				}
				args = append(args, &VeniceInteger{value})
			case TOKEN_STRING:
				args = append(args, &VeniceString{token.Value})
			default:
				log.Fatalf("Unexpected token: %q", token.Value)
			}

			token = lexer.NextToken()
		}

		bytecodes = append(bytecodes, &Bytecode{instruction.Value, args})
	}

	vm := NewVirtualMachine()
	_, err = vm.Execute(bytecodes)
	if err != nil {
		log.Fatalf("Execution error: %s", err)
	}
}
