package main

import (
	"bufio"
	"fmt"
	"github.com/shurcooL/go-goon"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

func main() {
	if len(os.Args) == 1 {
		repl()
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

func repl() {
	fmt.Print("The Venice programming language.\n\n")

	vm := NewVirtualMachine()
	compiler := NewCompiler()
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(">>> ")
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		operation := "execute"
		if line[0] == '!' {
			splitLine := strings.SplitN(line, " ", 2)
			if len(splitLine) > 0 {
				cmd := splitLine[0]
				line = splitLine[1]
				switch cmd {
				case "!compile":
					operation = "compile"
				case "!lex":
					operation = "lex"
				case "!parse":
					operation = "parse"
				case "!stack":
					if len(vm.stack) == 0 {
						fmt.Println("Stack is empty")
					} else {
						fmt.Println("Stack (bottom to top)")
						for _, value := range vm.stack {
							fmt.Println(value.Serialize())
						}
					}
					continue
				default:
					fmt.Printf("Error: unknown command %q\n", line)
					continue
				}
			}
		}

		lexer := NewLexer(line)
		if operation == "lex" {
			token := lexer.NextToken()
			for token.Type != TOKEN_EOF {
				fmt.Println(token.asString())
				token = lexer.NextToken()
			}
			continue
		}

		tree, err := NewParser(lexer).Parse()
		if err != nil {
			fmt.Printf("Parse error: %v\n", err)
			continue
		}

		if operation == "parse" {
			goon.Dump(tree)
			continue
		}

		code, err := compiler.Compile(tree)
		if err != nil {
			fmt.Printf("Compile error: %v\n", err)
			continue
		}

		if operation == "compile" {
			for _, bytecode := range code {
				fmt.Print(bytecode.Name)
				for _, arg := range bytecode.Args {
					fmt.Printf(" %s", arg.Serialize())
				}
				fmt.Print("\n")
			}
			continue
		}

		value, err := vm.Execute(code)
		if err != nil {
			fmt.Printf("Execution error: %v\n", err)
			continue
		}

		if value != nil {
			fmt.Println(value.Serialize())
		}
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

	code, err := NewCompiler().Compile(tree)
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

	WriteBytecodeListToFile(writer, code)
}

func execute_program(p string) {
	if strings.HasSuffix(p, ".vn") {
		log.Fatal("Error: can only execute compiled programs.")
	}

	data, err := ioutil.ReadFile(p)
	if err != nil {
		log.Fatal(err)
	}

	bytecodeList, err := ReadBytecodeListFromString(string(data))
	if err != nil {
		log.Fatal(err)
	}

	vm := NewVirtualMachine()
	_, err = vm.Execute(bytecodeList)
	if err != nil {
		log.Fatalf("Execution error: %s", err)
	}
}
