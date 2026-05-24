package main

import (
	"fmt"
	"os"

	"github.com/Grizak/Wick/src/backends"
	"github.com/Grizak/Wick/src/parser"
	"github.com/Grizak/Wick/src/tokenizer"
	"github.com/Grizak/Wick/src/tools"
	"github.com/Grizak/Wick/src/types"
	"github.com/alexflint/go-arg"
	randchars "github.com/mohae/randchars"
)

var version string // Filled in by ldflags at build time

type Args struct {
	Input              []string `arg:"positional,required" help:"Input file(s)"`
	Output             string   `arg:"-o,--output" help:"Output file"`
	SaveIntermediaries bool     `arg:"-s,--save-intermediaries" help:"Save intermediary files"`
	Target             string   `arg:"-t,--target" env:"WICK_TARGET" default:"linux/amd64" help:"Compilation target (default: linux/amd64)"`
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

	tools.Init()

	backend, err := backends.New(args.Target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", "failed to create backend", args.Target, err.Error())
		os.Exit(1)
	}

	for i := range args.Input {
		input := args.Input[i]

		content, err := os.ReadFile(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %s: %s\n", "failed to read input file", input, err.Error())
			os.Exit(1)
		}

		tokenizer := tokenizer.NewTokenizer(string(content))
		output := make(chan types.Token, 4096)
		tokenizer.Tokenize(output)

		parser := parser.NewParser()
		program := parser.Parse(output)

		outputFile := args.Output + "_" + string(randchars.LowerAlpha(8))

		backend.Generate(program, outputFile+".asm")

		backend.Assemble(outputFile+".asm", outputFile+".o", args.SaveIntermediaries)

		generatedFiles = append(generatedFiles, outputFile+".o")
	}

	backend.Link(generatedFiles, args.Output, args.SaveIntermediaries)
}
