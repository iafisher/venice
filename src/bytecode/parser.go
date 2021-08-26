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

	if len(compiledProgram.Imports) > 0 {
		for _, importObject := range compiledProgram.Imports {
			writer.WriteString(
				fmt.Sprintf(
					"import %q %q\n",
					importObject.Path,
					importObject.Name,
				),
			)
		}
		writer.WriteByte('\n')
	}

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
	compiledProgram, err := p.parse()
	if err != nil {
		return nil, err
	}

	err = p.resolveImports(compiledProgram)
	if err != nil {
		return nil, err
	}

	return compiledProgram, nil
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

	if p.currentToken.Type != lexer_mod.TOKEN_NEWLINE {
		p.newError("expected newline after version declaration")
	}

	p.nextTokenSkipNewlines()
	for p.currentToken.Type == lexer_mod.TOKEN_IMPORT {
		p.nextToken()
		path, _ := p.expectString()
		name, _ := p.expectString()

		if p.currentToken.Type != lexer_mod.TOKEN_NEWLINE {
			p.newError("expected newline after import statement")
		}
		p.nextTokenSkipNewlines()

		compiledProgram.Imports = append(
			compiledProgram.Imports, &CompiledProgramImport{Path: path, Name: name},
		)
	}

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
		} else if p.currentToken.Type == lexer_mod.TOKEN_DOUBLE_COLON {
			p.nextToken()
			symbolToken, ok := p.expect(lexer_mod.TOKEN_SYMBOL)
			if !ok {
				continue
			}
			p.currentFunctionName = fmt.Sprintf("%s::%s", symbol, symbolToken.Value)
			p.expect(lexer_mod.TOKEN_COLON)
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
			name, ok := p.expectString()
			if !ok {
				continue
			}

			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &BuildClass{name, n}
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
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &CallBuiltin{n}
		case "CALL_FUNCTION":
			n, ok := p.expectInt()
			if !ok {
				continue
			}

			bytecode = &CallFunction{n}
		case "CHECK_LABEL":
			name, ok := p.expectString()
			if !ok {
				continue
			}
			bytecode = &CheckLabel{name}
		case "FOR_ITER":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &ForIter{n}
		case "GET_ITER":
			bytecode = &GetIter{}
		case "LOOKUP_METHOD":
			name, ok := p.expectString()
			if !ok {
				continue
			}
			bytecode = &LookupMethod{name}
		case "PLACEHOLDER":
			p.newError("placeholder instruction")
			continue
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
		case "PUSH_CONST_FUNCTION":
			name, ok := p.expectString()
			if !ok {
				continue
			}

			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &PushConstFunction{name, n == 1}
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
		case "PUSH_ENUM_INDEX":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bytecode = &PushEnumIndex{n}
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
		case "STORE_INDEX":
			bytecode = &StoreIndex{}
		case "STORE_MAP_INDEX":
			bytecode = &StoreMapIndex{}
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

		compiledProgram.Code[p.currentFunctionName] = append(
			compiledProgram.Code[p.currentFunctionName],
			bytecode,
		)
		p.expect(lexer_mod.TOKEN_NEWLINE)

		if len(p.errors) > 0 {
			return nil, p.errors[0]
		}
	}

	if len(p.errors) == 0 {
		return compiledProgram, nil
	} else {
		return nil, p.errors[0]
	}
}

func (p *bytecodeParser) resolveImports(compiledProgram *CompiledProgram) error {
	visited := map[string]bool{}
	return p.resolveImportsRecursive(compiledProgram, compiledProgram, visited)
}

func (p *bytecodeParser) resolveImportsRecursive(
	original *CompiledProgram, compiledProgram *CompiledProgram, visited map[string]bool,
) error {
	for _, importObject := range compiledProgram.Imports {
		if _, ok := visited[importObject.Path]; ok {
			// TODO(2021-08-24): Better error message.
			return &BytecodeParseError{Message: "recursive imports", Location: nil}
		}

		// TODO(2021-08-24): This duplicates code in ReadCompiledProgramFromFile.
		fileContentsBytes, err := ioutil.ReadFile(importObject.Path)
		if err != nil {
			return err
		}

		subParser := bytecodeParser{
			lexer: lexer_mod.NewLexer(importObject.Path, string(fileContentsBytes)),
		}
		importedProgram, err := subParser.parse()
		if err != nil {
			return err
		}

		for functionName, functionCode := range importedProgram.Code {
			// TODO(2021-08-24): How should this be handled?
			if functionName == "main" {
				continue
			}

			qualifiedName := fmt.Sprintf("%s::%s", importObject.Name, functionName)
			original.Code[qualifiedName] = functionCode
		}

		visited[importObject.Path] = true
		p.resolveImportsRecursive(original, importedProgram, visited)
	}

	return nil
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
	if e.Location != nil {
		return fmt.Sprintf("%s at %s", e.Message, e.Location.String())
	} else {
		return e.Message
	}
}
