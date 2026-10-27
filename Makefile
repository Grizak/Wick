.PHONY: release snapshot build build-darwin clean check package dev

DEV_OUT ?= ./dist/wick
ROOT := $(CURDIR)
GOOS_HOST := $(shell go env GOOS)
GOARCH_HOST := $(shell go env GOARCH)
LLVM_INCLUDE := $(ROOT)/src/internal/llvm/toolchain/include/$(GOOS_HOST)_$(GOARCH_HOST)

# Lazily evaluated, so the script only runs for targets that need it
LDFLAGS_HOST = $(shell $(ROOT)/scripts/llvm-ldflags.sh $(GOOS_HOST)_$(GOARCH_HOST))
LDFLAGS_linux_amd64 = $(shell $(ROOT)/scripts/llvm-ldflags.sh linux_amd64)
LDFLAGS_linux_arm64 = $(shell $(ROOT)/scripts/llvm-ldflags.sh linux_arm64)
LDFLAGS_windows_amd64 = $(shell $(ROOT)/scripts/llvm-ldflags.sh windows_amd64)
LDFLAGS_darwin_arm64 = $(shell $(ROOT)/scripts/llvm-ldflags.sh darwin_arm64)

dev:
	mkdir -p $(dir $(DEV_OUT))
	mkdir -p src/internal/assets
	cp LICENSE src/internal/assets/LICENSE
	CC=gcc CXX=g++ \
	CGO_ENABLED=1 \
	CGO_CFLAGS="-I$(LLVM_INCLUDE)" \
	CGO_CXXFLAGS="-I$(LLVM_INCLUDE) -std=c++17" \
	CGO_LDFLAGS="$(LDFLAGS_HOST)" \
	go build -tags "byollvm llvm20" -ldflags "-X main.version=dev" -o "$(DEV_OUT)" ./src/cmd/wick

release:
	goreleaser release --clean

snapshot:
	goreleaser release --snapshot --clean

ZIG ?= /snap/bin/zig
WRAPPER_DIR := /tmp/zig-wrappers

# Builds Linux targets and Windows amd64 from Linux.
build:
	@command -v aarch64-linux-gnu-gcc >/dev/null && command -v aarch64-linux-gnu-g++ >/dev/null || { echo "Install g++-aarch64-linux-gnu to build linux/arm64" >&2; exit 1; }
	@mkdir -p $(WRAPPER_DIR)
	@echo '#!/bin/sh\nexec $(ZIG) cc -target x86_64-windows-gnu "$$@"' > $(WRAPPER_DIR)/x86_64-windows-gnu-cc
	@echo '#!/bin/sh\nexec $(ZIG) c++ -target x86_64-windows-gnu -stdlib=libstdc++ "$$@"' > $(WRAPPER_DIR)/x86_64-windows-gnu-c++
	@echo '#!/bin/sh\nexec $(ZIG) cc -target aarch64-windows-gnu "$$@"' > $(WRAPPER_DIR)/aarch64-windows-gnu-cc
	@echo '#!/bin/sh\nexec $(ZIG) c++ -target aarch64-windows-gnu -stdlib=libstdc++ "$$@"' > $(WRAPPER_DIR)/aarch64-windows-gnu-c++
	@chmod +x $(WRAPPER_DIR)/*
	PROJECT_ROOT=$(ROOT) PATH=$(WRAPPER_DIR):$$PATH \
	LLVM_LDFLAGS_LINUX_AMD64="$(LDFLAGS_linux_amd64)" \
	LLVM_LDFLAGS_LINUX_ARM64="$(LDFLAGS_linux_arm64)" \
	LLVM_LDFLAGS_WINDOWS_AMD64="$(LDFLAGS_windows_amd64)" \
	goreleaser build --snapshot --clean --id=linux-amd64 --id=linux-arm64 --id=windows-amd64

# Native macOS build - run this only on a Mac
build-darwin:
	CGO_ENABLED=1 \
	CGO_CFLAGS="-I$(ROOT)/src/internal/llvm/toolchain/include/darwin_arm64" \
	CGO_CXXFLAGS="-I$(ROOT)/src/internal/llvm/toolchain/include/darwin_arm64 -std=c++17" \
	CGO_LDFLAGS="$(LDFLAGS_darwin_arm64)" \
	go build -tags "byollvm llvm20" -o dist/wick-darwin-arm64 ./src/cmd/wick

check:
	goreleaser check

clean:
	rm -rf dist build

package:
	goreleaser release --skip=publish --clean

setup:
	git config core.hooksPath .githooks
