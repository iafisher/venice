package bytecode

import (
	"bufio"
	"fmt"
	lexer_mod "github.com/iafisher/venice/src/lexer"
	"io/ioutil"
	"strconv"
)

func WriteCompiledProgramToFile(writer *bufio.Writer, compiledProgram *CompiledProgram) {
	writer.WriteString(fmt.Sprintf("version %d\n\n", compiledProgram.Version))

	for functionName, functionCode := range compiledProgram.Code {
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

func ReadCompiledProgramFromFile(filePath string) (*CompiledProgram, error) {
	fileContentsBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	p := bytecodeParser{lexer: lexer_mod.NewLexer(filePath, string(fileContentsBytes))}
	return p.parse()
}

func ReadCompiledProgramFromString(programString string) (*CompiledProgram, error) {
	p := bytecodeParser{lexer: lexer_mod.NewLexer("", programString)}
	return p.parse()
}

type bytecodeParser struct {
	lexer               *lexer_mod.Lexer
	currentToken        *lexer_mod.Token
	currentFunctionName string
	errors              []error
}

func (p *bytecodeParser) parse() (*CompiledProgram, error) {
	compiledProgram := NewCompiledProgram()

	p.nextTokenSkipNewlines()

	if p.currentToken.Type != lexer_mod.TOKEN_SYMBOL && p.currentToken.Value != "version" {
		p.newError("expected version declaration at beginning of file")
	}

	p.nextToken()
	compiledProgram.Version, _ = p.expectInt()
	if len(p.errors) > 0 {
		return nil, p.errors[0]
	}

	for p.currentToken.Type != lexer_mod.TOKEN_EOF {
		if p.currentToken.Type == lexer_mod.TOKEN_NEWLINE {
			p.nextTokenSkipNewlines()
		}

		symbolToken, ok := p.expect(lexer_mod.TOKEN_SYMBOL)
		if !ok {
			continue
		}

		symbol := symbolToken.Value

		if p.currentToken.Type == lexer_mod.TOKEN_COLON {
			p.currentFunctionName = symbol
			p.nextToken()
			p.expect(lexer_mod.TOKEN_NEWLINE)
			continue
		}

		var bytecode Bytecode
		switch symbol {
		case "BINARY_ADD":
			bytecode = &BinaryAdd{}
		case "BINARY_AND":
			bytecode = &BinaryAnd{}
		case "BINARY_CONCAT":
			bytecode = &BinaryConcat{}
		case "BINARY_DIV":
			bytecode = &BinaryDiv{}
		case "BINARY_EQ":
			bytecode = &BinaryEq{}
		case "BINARY_GT":
			bytecode = &BinaryGt{}
		case "BINARY_GT_EQ":
			bytecode = &BinaryGtEq{}
		case "BINARY_IN":
			bytecode = &BinaryIn{}
		case "BINARY_LIST_INDEX":
			bytecode = &BinaryListIndex{}
		case "BINARY_LT":
			bytecode = &BinaryLt{}
		case "BINARY_LT_EQ":
			bytecode = &BinaryLtEq{}
		case "BINARY_MAP_INDEX":
			bytecode = &BinaryMapIndex{}
		case "BINARY_MUL":
			bytecode = &BinaryMul{}
		case "BINARY_NOT_EQ":
			bytecode = &BinaryNotEq{}
		case "BINARY_OR":
			bytecode = &BinaryOr{}
		case "BINARY_STRING_INDEX":
			bytecode = &BinaryStringIndex{}
		case "BINARY_SUB":
			bytecode = &BinarySub{}
		case "BUILD_CLASS":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &BuildClass{n}
		case "BUILD_LIST":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &BuildList{n}
		case "BUILD_MAP":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &BuildMap{n}
		case "BUILD_TUPLE":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &BuildTuple{n}
		case "CALL_BUILTIN":
			name, ok := p.expectString()
			if !ok {
				continue
			}
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &CallBuiltin{name, n}
		case "CALL_FUNCTION":
			name, ok := p.expectString()
			if !ok {
				continue
			}

			n, ok := p.expectInt()
			if !ok {
				continue
			}

			bytecode = &CallFunction{name, n}
		case "FOR_ITER":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &ForIter{n}
		case "GET_ITER":
			bytecode = &GetIter{}
		case "PUSH_CONST_BOOL":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &PushConstBool{n == 1}
		case "PUSH_CONST_CHAR":
			value, ok := p.expectString()
			if !ok {
				continue
			}
			bytecode = &PushConstChar{value[0]}
		case "PUSH_CONST_INT":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &PushConstInt{n}
		case "PUSH_CONST_STR":
			value, ok := p.expectString()
			if !ok {
				continue
			}
			bytecode = &PushConstStr{value}
		case "PUSH_ENUM":
			name, ok := p.expectString()
			if !ok {
				continue
			}
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &PushEnum{name, n}
		case "PUSH_FIELD":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &PushField{n}
		case "PUSH_NAME":
			name, ok := p.expectString()
			if !ok {
				continue
			}
			bytecode = &PushName{name}
		case "PUSH_TUPLE_FIELD":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &PushTupleField{n}
		case "REL_JUMP":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &RelJump{n}
		case "REL_JUMP_IF_FALSE":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &RelJumpIfFalse{n}
		case "REL_JUMP_IF_FALSE_OR_POP":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &RelJumpIfFalseOrPop{n}
		case "REL_JUMP_IF_TRUE_OR_POP":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &RelJumpIfTrueOrPop{n}
		case "RETURN":
			bytecode = &Return{}
		case "STORE_FIELD":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &StoreField{n}
		case "STORE_NAME":
			name, ok := p.expectString()
			if !ok {
				continue
			}
			bytecode = &StoreName{name}
		case "UNARY_MINUS":
			bytecode = &UnaryMinus{}
		case "UNARY_NOT":
			bytecode = &UnaryNot{}
		default:
			p.newError(fmt.Sprintf("unknown bytecode op `%s`", symbol))
			p.skipToNextLine()
			continue
		}

		compiledProgram.Code[p.currentFunctionName] = append(compiledProgram.Code[p.currentFunctionName], bytecode)
		p.expect(lexer_mod.TOKEN_NEWLINE)
	}

	if len(p.errors) == 0 {
		return compiledProgram, nil
	} else {
		return nil, p.errors[0]
	}
}

func (p *bytecodeParser) expectInt() (int, bool) {
	token, ok := p.expect(lexer_mod.TOKEN_INT)
	if !ok {
		return 0, false
	}

	n, err := strconv.ParseInt(token.Value, 0, 0)
	if err != nil {
		// TODO(2021-08-15): This will have the wrong location because the previous call
		// to `bytecodeParser.expect` advances past the integer token.
		p.newError("could not parse integer token")
		return 0, false
	}

	return int(n), true
}

func (p *bytecodeParser) expectString() (string, bool) {
	token, ok := p.expect(lexer_mod.TOKEN_STRING)
	if !ok {
		return "", false
	}
	return token.Value, true
}

func (p *bytecodeParser) expect(tokenType string) (*lexer_mod.Token, bool) {
	if p.currentToken.Type == tokenType {
		token := p.currentToken
		if tokenType == lexer_mod.TOKEN_NEWLINE {
			p.nextTokenSkipNewlines()
		} else {
			p.nextToken()
		}
		return token, true
	} else {
		actualType := p.currentToken.Type
		p.newError(fmt.Sprintf("expected %s, got %s", tokenType, actualType))
		p.skipToNextLine()
		return nil, false
	}
}

func (p *bytecodeParser) skipToNextLine() {
	for {
		if p.currentToken.Type == lexer_mod.TOKEN_EOF {
			break
		} else if p.currentToken.Type == lexer_mod.TOKEN_NEWLINE {
			p.nextTokenSkipNewlines()
			break
		} else {
			p.nextToken()
		}
	}
}

func (p *bytecodeParser) newError(message string) {
	p.errors = append(p.errors, &BytecodeParseError{message, p.currentToken.Location})
}

func (p *bytecodeParser) nextToken() *lexer_mod.Token {
	token := p.lexer.NextToken()
	p.currentToken = token
	return token
}

func (p *bytecodeParser) nextTokenSkipNewlines() *lexer_mod.Token {
	token := p.lexer.NextTokenSkipNewlines()
	p.currentToken = token
	return token
}

type BytecodeParseError struct {
	Message  string
	Location *lexer_mod.Location
}

func (e *BytecodeParseError) Error() string {
	return fmt.Sprintf("%s at %s", e.Message, e.Location.String())
}
