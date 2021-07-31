package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		repl_parser()
	} else if len(os.Args) == 2 {
		switch os.Args[1] {
		case "lex":
			repl_lexer()
		case "parse":
			repl_parser()
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
		lexer := NewLexer(line)
		parser := NewParser(lexer)
		expression, ok := parser.ParseExpression()
		if ok {
			fmt.Printf("%+v\n", expression)
		} else {
			fmt.Println("Parse error")
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
