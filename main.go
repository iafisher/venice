package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/iafisher/venice/src/common/bytecode"
	"github.com/iafisher/venice/src/common/lex"
	compilerPkg "github.com/iafisher/venice/src/compiler"
	"github.com/iafisher/venice/src/vm"
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

	virtualMachine := vm.NewVirtualMachine()
	compiler := compilerPkg.NewCompiler()
	scanner := bufio.NewScanner(os.Stdin)
	compiledProgram := bytecode.NewCompiledProgram()
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
					if len(virtualMachine.Stack) == 0 {
						fmt.Println("Stack is empty")
					} else {
						fmt.Println("Stack (bottom to top)")
						for _, value := range virtualMachine.Stack {
							fmt.Println(value.String())
						}
					}
					continue
				case "!symbols":
					for key, value := range virtualMachine.Env.Symbols {
						fmt.Printf("%s: %s\n", key, value.String())
					}
					continue
				case "!symbolTypes":
					compiler.PrintSymbolTable()
					continue
				case "!type":
					operation = "type"
				case "!types":
					compiler.PrintTypeSymbolTable()
					continue
				default:
					fmt.Printf("Error: unknown command %q\n", line)
					continue
				}
			}
		}

		lexer := lex.NewLexer("", line)
		if operation == "lex" {
			token := lexer.NextToken()
			for token.Type != lex.TOKEN_EOF {
				fmt.Println(token.String())
				token = lexer.NextToken()
			}
			continue
		}

		parsedFile, err := compilerPkg.NewParser().ParseString(line)
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

			parsedFile, err = compilerPkg.NewParser().ParseString(sb.String())
		}

		if err != nil {
			fmt.Printf("Parse error: %v\n", err)
			continue
		}

		if operation == "parse" {
			fmt.Println(parsedFile.String())
			continue
		}

		if operation == "type" {
			if len(parsedFile.Statements) == 0 {
				fmt.Println("Parse error: empty input")
				continue
			} else if len(parsedFile.Statements) > 1 {
				fmt.Println("Parse error: too many statements")
				continue
			}

			exprStmt, ok := parsedFile.Statements[0].(*compilerPkg.ExpressionStatementNode)
			if !ok {
				fmt.Println("Compile error: can only get type of expression, not statement")
				continue
			}

			exprType, err := compiler.GetType(exprStmt.Expr)
			if err != nil {
				fmt.Printf("Compile error: %v\n", err)
			} else {
				fmt.Println(exprType.String())
			}
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
				if functionName != "main" {
					continue
				}

				for _, bcode := range code {
					fmt.Printf("%s\n", bcode)
				}
			}
			continue
		}

		debug := (operation == "debug")
		value, err := virtualMachine.Execute(compiledProgram, debug)
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
	parsedFile, err := compilerPkg.NewParser().ParseFile(filePath)
	if err != nil {
		fatalError("Parse error: %v", err)
	}

	code, err := compilerPkg.NewCompiler().Compile(parsedFile)
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

	bytecode.WriteCompiledProgramToFile(writer, code)
}

func executeProgram(filePath string) {
	if strings.HasSuffix(filePath, ".vn") {
		compileProgram(filePath, false)
		filePath = filePath + "b"
	}

	compiledProgram, err := bytecode.ReadCompiledProgramFromFile(filePath)
	if err != nil {
		fatalError("Error while reading %s: %v", filePath, err)
	}

	virtualMachine := vm.NewVirtualMachine()
	_, err = virtualMachine.Execute(compiledProgram, false)
	if err != nil {
		fatalError("Execution error: %s", err)
	}
}

func fatalError(formatString string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, formatString+"\n", a...)
	os.Exit(1)
}
