package main

import (
	"fmt"
	"github.com/iafisher/venice/src/codegen"
	"github.com/iafisher/venice/src/compile"
	"github.com/iafisher/venice/src/parse"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "error: expected exactly one command-line argument")
		os.Exit(1)
	}

	filePath := os.Args[1]
	if strings.HasPrefix(filePath, "-") {
		fmt.Fprintf(os.Stderr, "error: %s accepts no command-line flags\n", os.Args[0])
		os.Exit(1)
	}

	compileProgram(filePath)
}

func compileProgram(filePath string) {
	ast, err := parse.Parse(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}

	program, err := compile.Compile(ast)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}

	assemblyOutputPath := getAssemblyOutputPath(filePath)
	err = codegen.GenerateCode(program, assemblyOutputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(2)
	}

	finalOutputPath := getFinalOutputPath(filePath)
	cmd := exec.Command("gcc", "-o", finalOutputPath, assemblyOutputPath)
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while running gcc: %s", err)
		os.Exit(2)
	}
}

func getAssemblyOutputPath(inputPath string) string {
	return replaceFileExtension(inputPath, ".s")
}

func getFinalOutputPath(inputPath string) string {
	return replaceFileExtension(inputPath, "")
}

func replaceFileExtension(path string, newExtension string) string {
	dotIndex := strings.LastIndexByte(path, '.')
	if dotIndex == -1 {
		return path + newExtension
	} else {
		return path[:dotIndex] + newExtension
	}
}
