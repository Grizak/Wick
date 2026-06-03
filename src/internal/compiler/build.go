package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/Grizak/Wick/src/internal/lexer"
	"github.com/Grizak/Wick/src/internal/llvm/assembler"
	"github.com/Grizak/Wick/src/internal/llvm/codegen"
	"github.com/Grizak/Wick/src/internal/llvm/linker"
	"github.com/Grizak/Wick/src/internal/llvm/toolchain"
	"github.com/Grizak/Wick/src/internal/parser"
	"github.com/Grizak/Wick/src/internal/target"
	"github.com/Grizak/Wick/src/internal/types"
	"github.com/mohae/randchars"
)

type BuildOptions struct {
	Input              []string
	Output             string
	SaveIntermediaries bool
	Target             string
	KeepOutput         bool
}

func Build(args BuildOptions) error {
	// Get path from Output and make sure that the parent directory exists
	outDir := filepath.Dir(args.Output)
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("Failed to create output directory %s: %v\n", outDir, err)
		}
	} else if !args.KeepOutput {
		// Clear output directory
		files, err := os.ReadDir(outDir)
		if err != nil {
			return fmt.Errorf("Failed to read output directory %s: %v\n", outDir, err)
		}
		for _, file := range files {
			if err := os.RemoveAll(filepath.Join(outDir, file.Name())); err != nil {
				return fmt.Errorf("Failed to clear output directory %s: %v\n", outDir, err)
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
		return err
	}

	// Make sure that the input files exists
	for i := range args.Input {
		input := args.Input[i]
		if _, err := os.Stat(input); os.IsNotExist(err) {
			return fmt.Errorf("Failed to read input file %s: %v\n", input, err)
		}
	}

	generatedFiles := make([]string, len(args.Input))

	toolchain.Init()

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
				return
			}

			outputFile := args.Output + "_" + string(randchars.LowerAlpha(8))

			generator := codegen.NewGenerator(&program)
			targetTriple, ok := target.TargetTriples[args.Target]
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

			if err := assembler.Assemble(outputFile+".ll", outputFile+".o", args.Output, args.SaveIntermediaries, index); err != nil {
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
		return fmt.Errorf("%d error(s) occurred during compilation", errCounter)
	}

	linker.Link(generatedFiles, args.Output, args.SaveIntermediaries, target.TargetTriples[args.Target])

	return nil
}

func validateTarget(host, target string) error {
	// Windows can only be targeted from Windows
	if strings.HasPrefix(target, "windows/") &&
		host != "windows" {
		return fmt.Errorf("cross-compiling to Windows is not currently supported")
	}
	return nil
}
