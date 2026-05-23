package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/mohae/randchars"
)

var version string

type TokenType string

const (
	TokenExit       TokenType = "exit"
	TokenOpenParen  TokenType = "("
	TokenCloseParen TokenType = ")"
	TokenIntLit     TokenType = "int_lit"
	TokenEOF        TokenType = "eof"
)

type Token struct {
	_type TokenType
	value *string
	line  int
}

type Args struct {
	Input              []string `arg:"positional,required" help:"Input file(s)"`
	Output             string   `arg:"-o,--output" help:"Output file"`
	SaveIntermediaries bool     `arg:"-s,--save-intermediaries" help:"Save intermediary files"`
}

func (Args) Version() string {
	return version
}

var args Args

// Parse args to get input file, then read input file, tokenize it, parse it, generate asm and write it to a file, pass it to nasm and ld
func main() {
	arg.MustParse(&args)

	// Make sure that the input files exists
	for i := range args.Input {
		input := args.Input[i]
		if _, err := os.Stat(input); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "%s: %s, %s\n", "Failed to read input file", input, "file doesn't exist")
			os.Exit(1)
		}
	}

	var generatedFiles []string

	for i := range args.Input {
		input := args.Input[i]

		content, err := os.ReadFile(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %s: %s\n", "failed to read input file", input, err.Error())
			os.Exit(1)
		}

		tokenizer := NewTokenizer(string(content))
		output := make(chan Token, 4096)
		tokenizer.Tokenize(output)

		parser := NewParser()
		program := parser.Parse(output)

		generator := NewGenerator(program)
		// Use the generator to generate assembly code
		asm := generator.Generate()

		// Write the assembly code to a file
		outputFile := args.Output
		if outputFile == "" {
			outputFile = input[:len(input)-len(".wi")]
		}

		outputFile += "_" + string(randchars.LowerAlpha(8)) + ".asm"

		generatedFiles = append(generatedFiles, outputFile[:len(outputFile)-len(".asm")]+".o")

		err = os.WriteFile(outputFile, []byte(asm), 0644)

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %s: %s\n", "failed to write output file", outputFile, err.Error())
			os.Exit(1)
		}

		if !args.SaveIntermediaries {
			defer os.Remove(outputFile)
		}

		// Use nasm to compile the assembly code to an object file
		objectFile := outputFile[:len(outputFile)-len(".asm")] + ".o"
		cmd := fmt.Sprintf("nasm -f elf64 %s -o %s", outputFile, objectFile)
		err = executeCommand(cmd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", "failed to compile assembly code", err.Error())
			os.Exit(1)
		}

		if !args.SaveIntermediaries {
			defer os.Remove(objectFile)
		}
	}

	// Use ld to link the object files into an executable
	cmd := fmt.Sprintf("ld -o %s %s", args.Output, strings.Join(generatedFiles, " "))
	err := executeCommand(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", "failed to link object files", err.Error())
		os.Exit(1)
	}
}

// executeCommand runs a shell command string and returns any error encountered.
func executeCommand(cmd string) error {
	c := exec.Command("/bin/sh", "-c", cmd)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
