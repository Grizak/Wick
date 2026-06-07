# Wick

A compiled, statically typed programming language with a clean, Python-inspired syntax that compiles to native executables via LLVM IR.

## Features

- Clean, minimal syntax
- Static typing with type inference
- Constant folding at compile time
- Scoped blocks
- `if`/`else` control flow
- `for` loops (infinite, while-style, and C-style)
- `break`/`continue`
- Compiles to native executables via LLVM IR
- Self-contained compiler binary — no external dependencies required
- Cross-compilation support

## Installation

Download the latest release for your platform from the [releases page](https://github.com/Grizak/Wick/releases).

## Usage

```bash
# Compile a file
wickc input.wi -o output

# Cross-compile to a specific target
wickc input.wi -o output -t linux/amd64

# Save intermediary files (.ll and .o)
wickc input.wi -o output -s
```

## Supported Hosts

| Host          | Supported |
| ------------- | --------- |
| Linux/amd64   | ✅        |
| macOS/arm64   | ✅        |
| Windows/amd64 | ✅        |

## Supported Targets

| Target        | From Linux | From macOS | From Windows |
| ------------- | ---------- | ---------- | ------------ |
| linux/amd64   | ✅         | ✅         | ✅           |
| linux/arm64   | ✅         | ✅         | ✅           |
| darwin/amd64  | ✅         | ✅         | ✅           |
| darwin/arm64  | ✅         | ✅         | ✅           |
| windows/amd64 | ❌         | ❌         | ✅           |
| windows/arm64 | ❌         | ❌         | ✅           |

## Syntax

### Comments

```wick
// This is a line comment
```

### Variables

```wick
// Mutable variable with type inference
let x = 42

// Mutable variable with explicit type
let x: int = 42

// Immutable constant
const x = 42
const x: int = 42

// Reassignment
x = 100
```

### Arithmetic

```wick
let x = 1 + 2
let y = x * 3 - (4 / 2)
```

### Control Flow

```wick
if x > 10 {
    // ...
} else if x > 5 {
    // ...
} else {
    // ...
}
```

### Loops

```wick
// Infinite loop
for {
    break
}

// While-style
for x > 0 {
    x = x - 1
}

// C-style
for let i = 0; i < 10; i = i + 1 {
    if i == 5 {
        continue
    }
}
```

### Exit

```wick
exit(0)
exit(x + 1)
```

## Building from Source

Requires Go 1.21+.

```bash
git clone https://github.com/Grizak/Wick
cd Wick
make
```

The compiler binary will be at `build/wick`.

## Roadmap

- [ ] Functions
- [ ] Built-in functions (`print`, etc.)
- [ ] More types (`bool`, `float`, `string`)
- [ ] Arrays
- [ ] Pointers

## License

MIT
