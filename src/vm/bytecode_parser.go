package vm

import (
	"bufio"
	"fmt"
	"github.com/iafisher/venice/src/common/bytecode"
	"github.com/iafisher/venice/src/common/lex"
	"io/ioutil"
	"strconv"
)

func WriteCompiledProgramToFile(writer *bufio.Writer, compiledProgram *bytecode.CompiledProgram) {
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
		for _, bcode := range functionCode {
			writer.WriteString("  ")
			writer.WriteString(bcode.String())
			writer.WriteByte('\n')
		}
	}

	writer.Flush()
}

func ReadCompiledProgramFromFile(filePath string) (*bytecode.CompiledProgram, error) {
	fileContentsBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	p := bytecodeParser{lexer: lex.NewLexer(filePath, string(fileContentsBytes))}
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

func ReadCompiledProgramFromString(programString string) (*bytecode.CompiledProgram, error) {
	p := bytecodeParser{lexer: lex.NewLexer("", programString)}
	return p.parse()
}

type bytecodeParser struct {
	lexer               *lex.Lexer
	currentToken        *lex.Token
	currentFunctionName string
	errors              []error
}

func (p *bytecodeParser) parse() (*bytecode.CompiledProgram, error) {
	compiledProgram := bytecode.NewCompiledProgram()

	p.nextTokenSkipNewlines()

	if p.currentToken.Type != lex.TOKEN_SYMBOL && p.currentToken.Value != "version" {
		p.newError("expected version declaration at beginning of file")
	}

	p.nextToken()
	compiledProgram.Version, _ = p.expectInt()

	if p.currentToken.Type != lex.TOKEN_NEWLINE {
		p.newError("expected newline after version declaration")
	}

	p.nextTokenSkipNewlines()
	for p.currentToken.Type == lex.TOKEN_IMPORT {
		p.nextToken()
		path, _ := p.expectString()
		name, _ := p.expectString()

		if p.currentToken.Type != lex.TOKEN_NEWLINE {
			p.newError("expected newline after import statement")
		}
		p.nextTokenSkipNewlines()

		compiledProgram.Imports = append(
			compiledProgram.Imports,
			&bytecode.CompiledProgramImport{Path: path, Name: name},
		)
	}

	if len(p.errors) > 0 {
		return nil, p.errors[0]
	}

	for p.currentToken.Type != lex.TOKEN_EOF {
		if p.currentToken.Type == lex.TOKEN_NEWLINE {
			p.nextTokenSkipNewlines()
		}

		symbolToken, ok := p.expect(lex.TOKEN_SYMBOL)
		if !ok {
			continue
		}

		symbol := symbolToken.Value

		if p.currentToken.Type == lex.TOKEN_COLON {
			p.currentFunctionName = symbol
			p.nextToken()
			p.expect(lex.TOKEN_NEWLINE)
			continue
		} else if p.currentToken.Type == lex.TOKEN_DOUBLE_COLON {
			p.nextToken()
			symbolToken, ok := p.expect(lex.TOKEN_SYMBOL)
			if !ok {
				continue
			}
			p.currentFunctionName = fmt.Sprintf("%s::%s", symbol, symbolToken.Value)
			p.expect(lex.TOKEN_COLON)
			continue
		}

		var bcode bytecode.Bytecode
		switch symbol {
		case "BINARY_ADD":
			bcode = &bytecode.BinaryAdd{}
		case "BINARY_AND":
			bcode = &bytecode.BinaryAnd{}
		case "BINARY_CONCAT":
			bcode = &bytecode.BinaryConcat{}
		case "BINARY_EQ":
			bcode = &bytecode.BinaryEq{}
		case "BINARY_GT":
			bcode = &bytecode.BinaryGt{}
		case "BINARY_GT_EQ":
			bcode = &bytecode.BinaryGtEq{}
		case "BINARY_IN":
			bcode = &bytecode.BinaryIn{}
		case "BINARY_LIST_INDEX":
			bcode = &bytecode.BinaryListIndex{}
		case "BINARY_LT":
			bcode = &bytecode.BinaryLt{}
		case "BINARY_LT_EQ":
			bcode = &bytecode.BinaryLtEq{}
		case "BINARY_MAP_INDEX":
			bcode = &bytecode.BinaryMapIndex{}
		case "BINARY_MUL":
			bcode = &bytecode.BinaryMul{}
		case "BINARY_NOT_EQ":
			bcode = &bytecode.BinaryNotEq{}
		case "BINARY_OR":
			bcode = &bytecode.BinaryOr{}
		case "BINARY_REAL_ADD":
			bcode = &bytecode.BinaryRealAdd{}
		case "BINARY_REAL_DIV":
			bcode = &bytecode.BinaryRealDiv{}
		case "BINARY_REAL_MUL":
			bcode = &bytecode.BinaryRealMul{}
		case "BINARY_REAL_SUB":
			bcode = &bytecode.BinaryRealSub{}
		case "BINARY_STRING_INDEX":
			bcode = &bytecode.BinaryStringIndex{}
		case "BINARY_SUB":
			bcode = &bytecode.BinarySub{}
		case "BUILD_CLASS":
			name, ok := p.expectString()
			if !ok {
				continue
			}

			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bcode = &bytecode.BuildClass{name, n}
		case "BUILD_LIST":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bcode = &bytecode.BuildList{n}
		case "BUILD_MAP":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bcode = &bytecode.BuildMap{n}
		case "BUILD_TUPLE":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bcode = &bytecode.BuildTuple{n}
		case "CALL_BUILTIN":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bcode = &bytecode.CallBuiltin{n}
		case "CALL_FUNCTION":
			n, ok := p.expectInt()
			if !ok {
				continue
			}

			bcode = &bytecode.CallFunction{n}
		case "CHECK_LABEL":
			name, ok := p.expectString()
			if !ok {
				continue
			}
			bcode = &bytecode.CheckLabel{name}
		case "FOR_ITER":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bcode = &bytecode.ForIter{n}
		case "GET_ITER":
			bcode = &bytecode.GetIter{}
		case "LOOKUP_METHOD":
			name, ok := p.expectString()
			if !ok {
				continue
			}
			bcode = &bytecode.LookupMethod{name}
		case "PLACEHOLDER":
			p.newError("placeholder instruction")
			continue
		case "PUSH_CONST_BOOL":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bcode = &bytecode.PushConstBool{n == 1}
		case "PUSH_CONST_FUNCTION":
			name, ok := p.expectString()
			if !ok {
				continue
			}

			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bcode = &bytecode.PushConstFunction{name, n == 1}
		case "PUSH_CONST_INT":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bcode = &bytecode.PushConstInt{n}
		case "PUSH_CONST_REAL_NUMBER":
			n, ok := p.expectRealNumber()
			if !ok {
				continue
			}
			bcode = &bytecode.PushConstRealNumber{n}
		case "PUSH_CONST_STR":
			value, ok := p.expectString()
			if !ok {
				continue
			}
			bcode = &bytecode.PushConstStr{value}
		case "PUSH_ENUM":
			name, ok := p.expectString()
			if !ok {
				continue
			}

			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bcode = &bytecode.PushEnum{name, n}
		case "PUSH_ENUM_INDEX":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bcode = &bytecode.PushEnumIndex{n}
		case "PUSH_FIELD":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bcode = &bytecode.PushField{n}
		case "PUSH_NAME":
			name, ok := p.expectString()
			if !ok {
				continue
			}
			bcode = &bytecode.PushName{name}
		case "PUSH_TUPLE_FIELD":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bcode = &bytecode.PushTupleField{n}
		case "REL_JUMP":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bcode = &bytecode.RelJump{n}
		case "REL_JUMP_IF_FALSE":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bcode = &bytecode.RelJumpIfFalse{n}
		case "REL_JUMP_IF_FALSE_OR_POP":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bcode = &bytecode.RelJumpIfFalseOrPop{n}
		case "REL_JUMP_IF_TRUE_OR_POP":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bcode = &bytecode.RelJumpIfTrueOrPop{n}
		case "RETURN":
			bcode = &bytecode.Return{}
		case "STORE_FIELD":
			n, ok := p.expectInt()
			if !ok {
				continue
			}
			bcode = &bytecode.StoreField{n}
		case "STORE_INDEX":
			bcode = &bytecode.StoreIndex{}
		case "STORE_MAP_INDEX":
			bcode = &bytecode.StoreMapIndex{}
		case "STORE_NAME":
			name, ok := p.expectString()
			if !ok {
				continue
			}
			bcode = &bytecode.StoreName{name}
		case "UNARY_MINUS":
			bcode = &bytecode.UnaryMinus{}
		case "UNARY_NOT":
			bcode = &bytecode.UnaryNot{}
		default:
			p.newError(fmt.Sprintf("unknown bytecode op `%s`", symbol))
			p.skipToNextLine()
			continue
		}

		compiledProgram.Code[p.currentFunctionName] = append(
			compiledProgram.Code[p.currentFunctionName],
			bcode,
		)
		p.expect(lex.TOKEN_NEWLINE)

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

func (p *bytecodeParser) resolveImports(compiledProgram *bytecode.CompiledProgram) error {
	visited := map[string]bool{}
	return p.resolveImportsRecursive(compiledProgram, compiledProgram, visited)
}

func (p *bytecodeParser) resolveImportsRecursive(
	original *bytecode.CompiledProgram,
	compiledProgram *bytecode.CompiledProgram,
	visited map[string]bool,
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
			lexer: lex.NewLexer(importObject.Path, string(fileContentsBytes)),
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
	token, ok := p.expect(lex.TOKEN_INT)
	if !ok {
		return 0, false
	}

	n, err := strconv.ParseInt(token.Value, 0, 0)
	if err != nil {
		// TODO(2021-08-15): This will have the wrong location because the previous call
		// to `bytecodeParser.expect` advances past the integer token.
		p.newError("invalid integer literal")
		return 0, false
	}

	return int(n), true
}

func (p *bytecodeParser) expectRealNumber() (float64, bool) {
	token, ok := p.expect(lex.TOKEN_REAL_NUMBER)
	if !ok {
		return 0, false
	}

	n, err := strconv.ParseFloat(token.Value, 64)
	if err != nil {
		// TODO(2021-08-15): This will have the wrong location because the previous call
		// to `bytecodeParser.expect` advances past the integer token.
		p.newError("invalid real number literal")
		return 0, false
	}

	return n, true
}

func (p *bytecodeParser) expectString() (string, bool) {
	token, ok := p.expect(lex.TOKEN_STRING)
	if !ok {
		return "", false
	}
	return token.Value, true
}

func (p *bytecodeParser) expect(tokenType string) (*lex.Token, bool) {
	if p.currentToken.Type == tokenType {
		token := p.currentToken
		if tokenType == lex.TOKEN_NEWLINE {
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
		if p.currentToken.Type == lex.TOKEN_EOF {
			break
		} else if p.currentToken.Type == lex.TOKEN_NEWLINE {
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

func (p *bytecodeParser) nextToken() *lex.Token {
	token := p.lexer.NextToken()
	p.currentToken = token
	return token
}

func (p *bytecodeParser) nextTokenSkipNewlines() *lex.Token {
	token := p.lexer.NextTokenSkipNewlines()
	p.currentToken = token
	return token
}

type BytecodeParseError struct {
	Message  string
	Location *lex.Location
}

func (e *BytecodeParseError) Error() string {
	if e.Location != nil {
		return fmt.Sprintf("%s at %s", e.Message, e.Location.String())
	} else {
		return e.Message
	}
}
