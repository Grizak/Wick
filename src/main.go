package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/Grizak/Wick/src/backend"
	"github.com/Grizak/Wick/src/generator"
	"github.com/Grizak/Wick/src/lexer"
	"github.com/Grizak/Wick/src/parser"
	"github.com/Grizak/Wick/src/tools"
	"github.com/Grizak/Wick/src/types"
	"github.com/alexflint/go-arg"
	randchars "github.com/mohae/randchars"
)

var version string // Filled in by ldflags at build time

type Args struct {
	Input              []string `arg:"positional,required" help:"Input file(s)"`
	Output             string   `arg:"-o,--output" default:"dist/out" help:"Output file"`
	SaveIntermediaries bool     `arg:"-s,--save-intermediaries" help:"Save intermediary files"`
	Target             string   `arg:"-t,--target" help:"Compilation target (default: GOOS/GOARCH, can also be set via WICK_TARGET environment variable)"`
	KeepOutput         bool     `arg:"-k,--keep-output" help:"Don't clear output directory when building"`
}

func (Args) Version() string {
	return version
}

var args Args

var outDir string

// Parse args to get input file, then read input file, tokenize it, parse it,
// generate llvm ir and write it to a file, pass it to llc and lld
func main() {
	arg.MustParse(&args)

	// Get path from Output and make sure that the parent directory exists
	outDir = filepath.Dir(args.Output)
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			fatal("Failed to create output directory %s: %v", outDir, err)
		}
	} else if !args.KeepOutput {
		// Clear output directory
		files, err := os.ReadDir(outDir)
		if err != nil {
			fatal("Failed to read output directory %s: %v", outDir, err)
		}
		for _, file := range files {
			if err := os.RemoveAll(filepath.Join(outDir, file.Name())); err != nil {
				fatal("Failed to clear output directory %s: %v", outDir, err)
			}
		}
	}

	// Determine target triple
	if args.Target == "" {
		args.Target = os.Getenv("WICK_TARGET")
	}

	if args.Target == "" {
		args.Target = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	}

	err := validateTarget(runtime.GOOS, args.Target)
	if err != nil {
		fatal("%s", err)
		os.Exit(1)
	}

	// Make sure that the input files exists
	for i := range args.Input {
		input := args.Input[i]
		if _, err := os.Stat(input); os.IsNotExist(err) {
			fatal("%s: %s, %s", "Failed to read input file", input, "file doesn't exist")
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

			lexer := lexer.NewLexer(string(content), input)
			output := make(chan types.LexerResult, 4096)
			go lexer.Tokenize(output)

			parser := parser.NewParser(filepath.Base(input))
			program, err := parser.Parse(output)
			if err != nil {
				results <- result{err: err}
			}

			outputFile := args.Output + "_" + string(randchars.LowerAlpha(8))

			generator := generator.NewGenerator(&program)
			targetTriple, ok := types.TargetTriples[args.Target]
			if !ok {
				results <- result{err: fmt.Errorf("unsupported target: %s", args.Target)}
				return
			}
			ir, err := generator.Generate(input, targetTriple)

			if err != nil {
				results <- result{err: err}
				return
			}

			if err := os.WriteFile(outputFile+".ll", []byte(ir), 0644); err != nil {
				results <- result{err: fmt.Errorf("failed to write LLVM IR to file for %s: %w", input, err)}
				return
			}

			if err := backend.Assemble(outputFile+".ll", outputFile+".o", args.Output, args.SaveIntermediaries, index); err != nil {
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

	errCounter := 0

	// Collect results and handle errors
	for r := range results {
		if r.err != nil {
			fmt.Fprintln(os.Stderr, r.err)
			errCounter++
		} else {
			generatedFiles[r.index] = r.outputFile
		}
	}

	if errCounter > 0 {
		os.Exit(1)
	}

	backend.Link(generatedFiles, args.Output, args.SaveIntermediaries, types.TargetTriples[args.Target])
}

func validateTarget(host, target string) error {
	// Windows can only be targeted from Windows
	if strings.HasPrefix(target, "windows/") &&
		host != "windows" {
		return fmt.Errorf("cross-compiling to Windows is not currently supported")
	}
	return nil
}

func fatal(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}
