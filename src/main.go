package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/Grizak/Wick/src/backend"
	"github.com/Grizak/Wick/src/generator"
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

var TargetTriples = map[string]string{
	"linux/amd64":   "x86_64-pc-linux-gnu",
	"linux/arm64":   "aarch64-pc-linux-gnu",
	"darwin/amd64":  "x86_64-apple-macosx",
	"darwin/arm64":  "aarch64-apple-macosx",
	"windows/amd64": "x86_64-pc-windows-msvc",
	"windows/arm64": "aarch64-pc-windows-msvc",
}

// Parse args to get input file, then read input file, tokenize it, parse it, generate llvm ir and write it to a file, pass it to llc and lld
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

	generatedFiles := make([]string, len(args.Input))

	tools.Init()

	type result struct {
		index      int
		outputFile string
		err        error
	}

	results := make(chan result, len(args.Input))
	var wg sync.WaitGroup

	for i := range args.Input {
		wg.Add(1)
		go func(input string, index int) {
			defer wg.Done()

			content, err := os.ReadFile(input)
			if err != nil {
				results <- result{err: fmt.Errorf("failed to read input file %s: %w", input, err)}
				return
			}

			tokenizer := tokenizer.NewTokenizer(string(content))
			output := make(chan types.Token, 4096)
			go tokenizer.Tokenize(output)

			parser := parser.NewParser()
			program := parser.Parse(output)

			outputFile := args.Output + "_" + string(randchars.LowerAlpha(8))

			generator := generator.NewGenerator(&program)
			targetTriple, ok := types.TargetTriples[args.Target]
			if !ok {
				results <- result{err: fmt.Errorf("unsupported target: %s", args.Target)}
				return
			}
			ir := generator.Generate(input, targetTriple)

			if err := os.WriteFile(outputFile+".ll", []byte(ir), 0644); err != nil {
				results <- result{err: fmt.Errorf("failed to write LLVM IR to file for %s: %w", input, err)}
				return
			}

			if err := backend.Assemble(outputFile+".ll", outputFile+".o", args.SaveIntermediaries); err != nil {
				results <- result{err: fmt.Errorf("assemble failed for %s: %w", input, err)}
				return
			}

			results <- result{outputFile: outputFile + ".o", index: index}
		}(args.Input[i], i)
	}

	// Close results once all goroutines finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results and handle errors
	for r := range results {
		if r.err != nil {
			fmt.Fprintln(os.Stderr, r.err)
			os.Exit(1)
		}
		generatedFiles[r.index] = r.outputFile
	}

	backend.Link(generatedFiles, args.Output, args.SaveIntermediaries, types.TargetTriples[args.Target])
}
