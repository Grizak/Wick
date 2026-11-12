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
- Namespaced modules
- Compiles to native executables via LLVM IR
- Self-contained compiler binary — no external dependencies required
- Cross-compilation support

## Installation

Download the latest release for your platform from the [releases page](https://github.com/Grizak/Wick/releases).

You can also download it from my official apt/rpm/arch repository, more info [here](https://pkg.isaksweb.xyz).

If you're on windows, it might say "missing dll dependency" or "The program cannot execute because [some filename] couldn't be found", in that case, you should install the [Microsoft Visual C++ Redistributable for Visual Studio](https://learn.microsoft.com/en-us/cpp/windows/latest-supported-vc-redist?view=msvc-170) (latest version) and try again.

## Usage

```bash
# Compile a file
wick input.wi -o output

# Cross-compile to a specific target
wick input.wi -o output -t linux/amd64

# Save intermediary files (.ll, .asm and .o)
wick input.wi -o output -s
```

## Supported Hosts

| Host          | Supported |
| ------------- | --------- |
| Linux/amd64   | ✅        |
| Linux/arm64   | ✅        |
| macOS/arm64   | ✅        |
| Windows/amd64 | ✅        |
| Windows/arm64 | ✅        |

## Supported Targets

This readme might not be up to date, run `wick version` to see full list of targets, they should be cross-compilable from all platforms

| Target        | From Linux | From macOS | From Windows |
| ------------- | ---------- | ---------- | ------------ |
| linux/amd64   | ✅         | ✅         | ✅           |
| linux/arm64   | ✅         | ✅         | ✅           |
| darwin/amd64  | ✅         | ✅         | ✅           |
| darwin/arm64  | ✅         | ✅         | ✅           |
| windows/amd64 | ✅         | ✅         | ✅           |
| windows/arm64 | ✅         | ✅         | ✅           |

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

### Modules

Import a Wick source file and call its functions through the filename-derived
namespace:

```wick
import "math.wk"

let result = math.add(2, 3)
```

Import paths are resolved relative to the importing file. The namespace is the
imported filename without its extension. Imported files may contain function
declarations and further imports; top-level executable statements and variables
are not supported in imported modules. Imports are only allowed at the top level.

## Building from Source

Requires Go 1.26.2+.

```bash
git clone https://github.com/Grizak/Wick
cd Wick
make setup
make
```

The compiler binary will be at `dist/wick`.

## Roadmap

- [ ] Libc + stdlib
- [ ] More types (`bool`, `string`)
- [ ] Arrays
- [ ] Pointers

## License

MIT
