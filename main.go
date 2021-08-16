package main

import (
	"bufio"
	"flag"
	"fmt"
	bytecode_mod "github.com/iafisher/venice/src/bytecode"
	"github.com/iafisher/venice/src/compiler"
	lexer_mod "github.com/iafisher/venice/src/lexer"
	"github.com/iafisher/venice/src/parser"
	vm_mod "github.com/iafisher/venice/src/vm"
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

	vm := vm_mod.NewVirtualMachine()
	compiler := compiler.NewCompiler()
	scanner := bufio.NewScanner(os.Stdin)
	compiledProgram := bytecode_mod.NewCompiledProgram()
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

		if line[len(line)-1] == '\\' {
			var sb strings.Builder
			sb.WriteString(line[:len(line)-1])
			sb.WriteByte('\n')
			for {
				fmt.Print("... ")
				ok := scanner.Scan()
				if !ok {
					return
				}

				nextLine := scanner.Text()
				nextLine = strings.TrimSpace(nextLine)
				if len(nextLine) == 0 {
					break
				}
				sb.WriteString(nextLine)
				sb.WriteByte('\n')
			}
			line = sb.String()
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
					if len(vm.Stack) == 0 {
						fmt.Println("Stack is empty")
					} else {
						fmt.Println("Stack (bottom to top)")
						for _, value := range vm.Stack {
							fmt.Println(value.String())
						}
					}
					continue
				case "!symbols":
					for key, value := range vm.Env.Symbols {
						fmt.Printf("%s: %s\n", key, value.String())
					}
					continue
				case "!symbolTypes":
					for key, value := range compiler.SymbolTable.Symbols {
						fmt.Printf("%s: %s\n", key, value.String())
					}
					continue
				case "!types":
					for key, value := range compiler.TypeSymbolTable.Symbols {
						fmt.Printf("%s: %s\n", key, value.String())
					}
					continue
				default:
					fmt.Printf("Error: unknown command %q\n", line)
					continue
				}
			}
		}

		lexer := lexer_mod.NewLexer("", line)
		if operation == "lex" {
			token := lexer.NextToken()
			for token.Type != lexer_mod.TOKEN_EOF {
				fmt.Println(token.String())
				token = lexer.NextToken()
			}
			continue
		}

		parsedFile, err := parser.NewParser().ParseString(line)
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

			parsedFile, err = parser.NewParser().ParseString(sb.String())
		}

		if err != nil {
			fmt.Printf("Parse error: %v\n", err)
			continue
		}

		if operation == "parse" {
			fmt.Println(parsedFile.String())
			continue
		}

		thisCompiledProgram, err := compiler.Compile(parsedFile)
		if err != nil {
			fmt.Printf("Compile error: %v\n", err)
			continue
		}

		for functionName, functionCode := range thisCompiledProgram.Code {
			compiledProgram.Code[functionName] = functionCode
		}

		if operation == "compile" {
			for functionName, code := range compiledProgram.Code {
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
			fmt.Println(value.String())
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
	parsedFile, err := parser.NewParser().ParseFile(filePath)
	if err != nil {
		fatalError("Parse error: %v", err)
	}

	code, err := compiler.NewCompiler().Compile(parsedFile)
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

	bytecode_mod.WriteCompiledProgramToFile(writer, code)
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
	bytecodeList, err := bytecode_mod.ReadCompiledProgramFromString(fileContents)
	if err != nil {
		fatalError("Error while reading %s: %v", filePath, err)
	}

	vm := vm_mod.NewVirtualMachine()
	_, err = vm.Execute(bytecodeList, false)
	if err != nil {
		fatalError("Execution error: %s", err)
	}
}

func fatalError(formatString string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, formatString+"\n", a...)
	os.Exit(1)
}
