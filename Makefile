.PHONY: release snapshot build clean check package dev

DEV_OUT ?= ./dist/wick
GOOS_HOST := $(shell go env GOOS)
GOARCH_HOST := $(shell go env GOARCH)
LLVM_TOOLCHAIN := $(CURDIR)/src/internal/llvm/toolchain
LLVM_INCLUDE := $(LLVM_TOOLCHAIN)/include/$(GOOS_HOST)_$(GOARCH_HOST)
LLVM_LIBDIR := $(LLVM_TOOLCHAIN)/lib/$(GOOS_HOST)_$(GOARCH_HOST)

ifeq ($(GOOS_HOST),windows)
  LLVM_LIB := $(LLVM_LIBDIR)/LLVM.lib
  EXTRA_LDFLAGS :=
else ifeq ($(GOOS_HOST),darwin)
  LLVM_LIB := $(LLVM_LIBDIR)/libLLVM.a
  EXTRA_LDFLAGS :=
else
  LLVM_LIB := $(LLVM_LIBDIR)/libLLVM.a
  # static c++ runtime, keep libc dynamic
  EXTRA_LDFLAGS := -static-libstdc++ -static-libgcc -lm -ldl -pthread
endif

dev:
	mkdir -p $(dir $(DEV_OUT))
	mkdir -p src/internal/assets
	cp LICENSE src/internal/assets/LICENSE
	CC=gcc CXX=g++ \
	CGO_ENABLED=1 \
	CGO_CFLAGS="-I$(LLVM_INCLUDE)" \
	CGO_CXXFLAGS="-I$(LLVM_INCLUDE) -std=c++17" \
	CGO_LDFLAGS="$(LLVM_LIB) $(EXTRA_LDFLAGS)" \
	go build -tags "byollvm llvm20" -ldflags "-X main.version=dev" -o "$(DEV_OUT)" ./src/cmd/wick

release:
	goreleaser release --clean

snapshot:
	goreleaser release --snapshot --clean

ZIG ?= /snap/bin/zig
WRAPPER_DIR := /tmp/zig-wrappers
ROOT := $(shell pwd)

# Builds Linux targets and Windows amd64 from Linux.
build:
	@command -v aarch64-linux-gnu-gcc >/dev/null && command -v aarch64-linux-gnu-g++ >/dev/null || { echo "Install g++-aarch64-linux-gnu to build linux/arm64" >&2; exit 1; }
	@mkdir -p $(WRAPPER_DIR)
	@echo '#!/bin/sh\nexec $(ZIG) cc -target x86_64-windows-gnu "$$@"' > $(WRAPPER_DIR)/x86_64-windows-gnu-cc
	@echo '#!/bin/sh\nexec $(ZIG) c++ -target x86_64-windows-gnu -stdlib=libstdc++ "$$@"' > $(WRAPPER_DIR)/x86_64-windows-gnu-c++
	@echo '#!/bin/sh\nexec $(ZIG) cc -target aarch64-windows-gnu "$$@"' > $(WRAPPER_DIR)/aarch64-windows-gnu-cc
	@echo '#!/bin/sh\nexec $(ZIG) c++ -target aarch64-windows-gnu -stdlib=libstdc++ "$$@"' > $(WRAPPER_DIR)/aarch64-windows-gnu-c++
	@chmod +x $(WRAPPER_DIR)/*
	PROJECT_ROOT=$(ROOT) PATH=$(WRAPPER_DIR):$$PATH goreleaser build --snapshot --clean --id=linux-amd64 --id=linux-arm64 --id=windows-amd64

# Native macOS build - run this only on a Mac
build-darwin:
	CGO_ENABLED=1 \
	CGO_CFLAGS="-I$(ROOT)/src/internal/llvm/toolchain/include/darwin_arm64" \
	CGO_CXXFLAGS="-I$(ROOT)/src/internal/llvm/toolchain/include/darwin_arm64 -std=c++17" \
	CGO_LDFLAGS="$(ROOT)/src/internal/llvm/toolchain/lib/darwin_arm64/libLLVM.a" \
	go build -tags "byollvm llvm20" -o dist/wick-darwin-arm64 ./src/cmd/wick

check:
	goreleaser check

clean:
	rm -rf dist build

package:
	goreleaser release --skip=publish --clean