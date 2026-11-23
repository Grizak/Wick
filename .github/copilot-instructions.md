# Wick Compiler

## Project overview

Wick is a compiled programming language implemented in Go.
The compiler translates Wick source code into LLVM IR,
then uses linked LLVM libraries to produce
native executables.

## Compiler architecture

The compilation pipeline is:

1. Lexer: source code -> tokens
2. Parser: tokens -> AST
3. Semantic analysis: AST validation and type checking
4. Code generation: AST -> LLVM IR
5. Optimization: LLVM IR -> optimized LLVM IR (if the user specifies the `-O` flag with a number between 1 and 3 (0, which is the default, is also valid but doesn't optimize))
6. Assembly: optimized LLVM IR -> targetMachine.EmitToMemoryBuffer -> object file
7. Linking: object file -> executable using linked lld

## Project structure

- src/cmd/wick: CLI entry point
- src/internal/ast: AST definitions
- src/internal/compiler: compiler orchestration
- src/internal/lexer: lexical analysis
- src/internal/parser: parsing
- src/internal/semantic: semantic analysis
- src/internal/llvm: LLVM integration
- src/internal/target: target platform handling
- src/internal/typesys: type system

## Development conventions

- Use Go for compiler implementation.
- Follow idiomatic Go conventions.
- Avoid unnecessary dependencies.
- Keep compiler stages separated.
- Prefer modifying existing abstractions over introducing
  duplicate implementations.
- Preserve existing CLI behavior unless explicitly requested.
- Update the changelog, the readme and docs/grammar.md whenever applicable.

## LLVM

- LLVM is used as the compiler backend.
- opt handles IR optimization.
- llc handles native code generation.
- lld handles linking.
- Respect the project's supported target architectures.

## Important

- Inspect existing implementations before making architectural changes.
- Don't assume that LLVM tools are installed globally.
- Check existing tests before changing compiler behavior.
- Explain significant architectural changes.

## Tests

Because this compiler depends on the llvm toolchain in src/internal/llvm/toolchain/lib, tests are run using the Makefile's test target in the root of the project. The Makefile sets up the environment and runs tests with `go test ./src/...`.

- Write tests for all new features and bug fixes.
- Write tests for edge cases and error conditions.
- Use the `testing` package for unit tests.
