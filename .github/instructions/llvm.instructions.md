---
applyTo: "src/internal/llvm/**/*.go"
---

# ALWAYS read .github/copilot-instructions.md BEFORE reading this file

# LLVM integration

- Follow the existing LLVM toolchain abstraction.
- Never assume LLVM binaries are available in PATH.
- Support the platforms already supported by Wick.
- Keep platform-specific logic isolated.
