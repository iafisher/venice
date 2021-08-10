package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/shurcooL/go-goon"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func main() {
	compileCommand := flag.NewFlagSet("compile", flag.ExitOnError)
	executeCommand := flag.NewFlagSet("execute", flag.ExitOnError)

	stdoutFlagPtr := compileCommand.Bool("stdout", false, "Output bytecode to stdout instead of to disk.")

	if len(os.Args) == 1 {
		repl()
	} else {
		switch os.Args[1] {
		case "compile":
			compileCommand.Parse(os.Args[2:])
			compileProgram(compileCommand.Arg(0), *stdoutFlagPtr)
		case "execute":
			executeCommand.Parse(os.Args[2:])
			executeProgram(executeCommand.Arg(0))
		default:
			fmt.Printf("Error: unknown subcommand %q", os.Args[1])
		}
	}
}

func repl() {
	fmt.Print("The Venice programming language.\n\n")
	fmt.Print("Type !help to view available commands.\n\n\n")

	vm := NewVirtualMachine()
	compiler := NewCompiler()
	scanner := bufio.NewScanner(os.Stdin)
	compiledProgram := NewCompiledProgram()
	for {
		fmt.Print(">>> ")
		ok := scanner.Scan()
		if !ok {
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
				if len(splitLine) > 1 {
					line = splitLine[1]
				}

				switch cmd {
				case "!compile":
					operation = "compile"
				case "!debug":
					operation = "debug"
				case "!help":
					fmt.Println(helpString)
					continue
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
				case "!symbols":
					for key, value := range vm.env.symbols {
						fmt.Printf("%s: %s\n", key, value.SerializePrintable())
					}
					continue
				case "!symbolTypes":
					for key, value := range compiler.symbolTable.symbols {
						fmt.Printf("%s: %s\n", key, value.String())
					}
					continue
				case "!types":
					for key, value := range compiler.typeSymbolTable.symbols {
						fmt.Printf("%s: %s\n", key, value.String())
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
		if err != nil && strings.HasPrefix(err.Error(), "premature end of input") {
			var sb strings.Builder
			sb.WriteString(line)
			sb.WriteByte('\n')
			for {
				fmt.Print("... ")
				ok := scanner.Scan()
				if !ok {
					return
				}

				nextLine := scanner.Text()
				nextLine = strings.TrimSpace(nextLine)
				sb.WriteString(nextLine)
				sb.WriteByte('\n')
				if len(nextLine) == 0 {
					break
				}
			}

			tree, err = NewParser(NewLexer(sb.String())).Parse()
		}

		if err != nil {
			fmt.Printf("Parse error: %v\n", err)
			continue
		}

		if operation == "parse" {
			goon.Dump(tree)
			continue
		}

		thisCompiledProgram, err := compiler.Compile(tree)
		if err != nil {
			fmt.Printf("Compile error: %v\n", err)
			continue
		}

		for functionName, functionCode := range thisCompiledProgram {
			compiledProgram[functionName] = functionCode
		}

		if operation == "compile" {
			for functionName, code := range compiledProgram {
				fmt.Printf("%s:\n", functionName)
				for _, bytecode := range code {
					fmt.Printf("  %s\n", bytecode)
				}
			}
			continue
		}

		debug := (operation == "debug")
		value, err := vm.Execute(compiledProgram, debug)
		if err != nil {
			fmt.Printf("Execution error: %v\n", err)
			continue
		}

		if value != nil {
			fmt.Println(value.Serialize())
		}
	}
}

const helpString = `!compile <code>   Compile the Venice code into bytecode.
!debug <code>     Execute the Venice code with debugging messages turned on.
!help             Print this help message.
!lex <code>       Lex the Venice code and print the resulting tokens.
!parse <code>     Parse the Venice code and print the resulting syntax tree.
!stack            Print the current state of the virtual machine stack.
!symbols          Print all symbols in the current environment and their values.
!symbolTypes      Print all symbols in the current environment and their types.
!types            Print all types in the current environment.`

func compileProgram(filePath string, toStdout bool) {
	fileContentsBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		fatalError("Error while opening %s: %v", filePath, err)
	}

	fileContents := string(fileContentsBytes)
	tree, err := NewParser(NewLexer(fileContents)).Parse()
	if err != nil {
		fatalError("Parse error: %v", err)
	}

	code, err := NewCompiler().Compile(tree)
	if err != nil {
		fatalError("Compile error: %v", err)
	}

	fileExt := path.Ext(filePath)
	outputPath := filePath[:len(filePath)-len(fileExt)] + ".vnb"

	var writer *bufio.Writer
	if toStdout {
		writer = bufio.NewWriter(os.Stdout)
	} else {
		f, err := os.Create(outputPath)
		if err != nil {
			fatalError("Error while creating %s: %v", outputPath, err)
		}
		defer f.Close()
		writer = bufio.NewWriter(f)
	}

	WriteCompiledProgramToFile(writer, code)
}

func executeProgram(filePath string) {
	if strings.HasSuffix(filePath, ".vn") {
		compileProgram(filePath, false)
		filePath = filePath + "b"
	}

	fileContentsBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		fatalError("Error while opening %s: %v", filePath, err)
	}

	fileContents := string(fileContentsBytes)
	bytecodeList, err := ReadCompiledProgramFromString(fileContents)
	if err != nil {
		fatalError("Error while reading %s: %v", filePath, err)
	}

	vm := NewVirtualMachine()
	_, err = vm.Execute(bytecodeList, false)
	if err != nil {
		fatalError("Execution error: %s", err)
	}
}

func fatalError(formatString string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, formatString+"\n", a...)
	os.Exit(1)
}
