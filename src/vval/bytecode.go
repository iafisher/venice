package vval

import (
	"bufio"
	"fmt"
	lexer_mod "github.com/iafisher/venice/src/lexer"
	"io/ioutil"
	"strconv"
	"strings"
)

type CompiledProgram map[string][]*Bytecode

func NewCompiledProgram() CompiledProgram {
	return map[string][]*Bytecode{"main": []*Bytecode{}}
}

type Bytecode struct {
	Name string
	Args []VeniceValue
}

func NewBytecode(name string, args ...VeniceValue) *Bytecode {
	return &Bytecode{name, args}
}

func (b *Bytecode) String() string {
	var sb strings.Builder
	sb.WriteString(b.Name)
	for _, arg := range b.Args {
		sb.WriteString(" ")
		sb.WriteString(arg.String())
	}
	return sb.String()
}

func WriteCompiledProgramToFile(writer *bufio.Writer, compiledProgram CompiledProgram) {
	for functionName, functionCode := range compiledProgram {
		writer.WriteString(functionName)
		writer.WriteString(":\n")
		for _, bytecode := range functionCode {
			writer.WriteString("  ")
			writer.WriteString(bytecode.String())
			writer.WriteByte('\n')
		}
	}

	writer.Flush()
}

func ReadCompiledProgramFromFile(filePath string) (CompiledProgram, error) {
	fileContentsBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return readCompiledProgramGeneric(filePath, string(fileContentsBytes))
}

func ReadCompiledProgramFromString(programString string) (CompiledProgram, error) {
	return readCompiledProgramGeneric("<string>", programString)
}

func readCompiledProgramGeneric(filePath string, programString string) (CompiledProgram, error) {
	compiledProgram := NewCompiledProgram()
	var currentFunctionName string
	for i, line := range strings.Split(programString, "\n") {
		lexer := lexer_mod.NewLexer(filePath, line)
		firstToken := lexer.NextToken()
		if firstToken.Type == lexer_mod.TOKEN_EOF {
			continue
		}

		if firstToken.Type != lexer_mod.TOKEN_SYMBOL {
			return nil, &BytecodeParseError{fmt.Sprintf("could not parse line %d", i+1)}
		}

		token := lexer.NextToken()
		if token.Type == lexer_mod.TOKEN_COLON {
			currentFunctionName = firstToken.Value
			continue
		}

		if currentFunctionName == "" {
			return nil, &BytecodeParseError{"bytecode instruction outside of function"}
		}

		args := []VeniceValue{}
		for token.Type != lexer_mod.TOKEN_EOF && token.Type != lexer_mod.TOKEN_NEWLINE {
			switch token.Type {
			case lexer_mod.TOKEN_CHARACTER:
				args = append(args, &VeniceCharacter{token.Value[0]})
			case lexer_mod.TOKEN_FALSE:
				args = append(args, &VeniceBoolean{false})
			case lexer_mod.TOKEN_INT:
				value, err := strconv.ParseInt(token.Value, 10, 0)
				if err != nil {
					return nil, &BytecodeParseError{"could not parse integer token"}
				}
				args = append(args, &VeniceInteger{int(value)})
			case lexer_mod.TOKEN_STRING:
				args = append(args, &VeniceString{token.Value})
			case lexer_mod.TOKEN_TRUE:
				args = append(args, &VeniceBoolean{true})
			default:
				return nil, &BytecodeParseError{fmt.Sprintf("unexpected token: %q", token.Value)}
			}

			token = lexer.NextToken()
		}

		compiledProgram[currentFunctionName] = append(compiledProgram[currentFunctionName], &Bytecode{firstToken.Value, args})
	}

	return compiledProgram, nil
}

type BytecodeParseError struct {
	Message string
}

func (e *BytecodeParseError) Error() string {
	return e.Message
}
