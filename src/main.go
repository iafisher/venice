package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		main_lexer_repl()
	} else if len(os.Args) == 2 {
		if os.Args[1] == "lex" {
			main_lexer_repl()
		} else {
			fmt.Printf("Error: unknown subcommand %q", os.Args[1])
		}
	} else {
		fmt.Println("Error: too many command-line arguments.")
	}
}

func main_lexer_repl() {
	fmt.Print("The Venice programming language.\n\n")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(">>> ")
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		lexer := NewLexer(line)
		token := lexer.NextToken()
		for token.Type != TOKEN_EOF {
			fmt.Println(token.asString())
			token = lexer.NextToken()
		}
	}
}
