.PHONY: release snapshot build build-darwin clean check package dev

DEV_OUT ?= ./dist/wick
ROOT := $(CURDIR)
GOOS_HOST := $(shell go env GOOS)
GOARCH_HOST := $(shell go env GOARCH)
LLVM_INCLUDE := $(ROOT)/src/internal/llvm/toolchain/include/$(GOOS_HOST)_$(GOARCH_HOST)
MUSL_TOOLCHAIN := $(ROOT)/toolchains/musl/linux_amd64
MUSL_CC := $(MUSL_TOOLCHAIN)/bin/x86_64-linux-musl-gcc
MUSL_CXX := $(MUSL_TOOLCHAIN)/bin/x86_64-linux-musl-g++

# Lazily evaluated, so the script only runs for targets that need it
LDFLAGS_HOST = $(shell $(ROOT)/scripts/llvm-ldflags.sh $(GOOS_HOST)_$(GOARCH_HOST))
LDFLAGS_linux_amd64 = $(shell $(ROOT)/scripts/llvm-ldflags.sh linux_amd64)
LDFLAGS_linux_arm64 = $(shell $(ROOT)/scripts/llvm-ldflags.sh linux_arm64)
LDFLAGS_windows_amd64 = $(shell $(ROOT)/scripts/llvm-ldflags.sh windows_amd64)
LDFLAGS_windows_arm64 = $(shell $(ROOT)/scripts/llvm-ldflags.sh windows_arm64)
LDFLAGS_darwin_arm64 = $(shell $(ROOT)/scripts/llvm-ldflags.sh darwin_arm64)

version := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

dev:
	mkdir -p $(dir $(DEV_OUT))
	go run ./scripts/prepare-assets.go
	CC=$(MUSL_CC) \
	CXX=$(MUSL_CXX) \
	CGO_ENABLED=1 \
	CGO_CFLAGS="-I$(LLVM_INCLUDE) -ffunction-sections -fdata-sections -flto" \
	CGO_CXXFLAGS="-I$(LLVM_INCLUDE) -std=c++17 -ffunction-sections -fdata-sections -flto" \
	CGO_LDFLAGS="$(LDFLAGS_HOST)" \
	go build -tags "byollvm llvm20" \
	-ldflags "-X main.version=$(version) -linkmode=external -extldflags=-static" \
	-o "$(DEV_OUT)" ./src/cmd/wick

release:
	goreleaser release --clean

snapshot:
	goreleaser release --snapshot --clean

check:
	goreleaser check

clean:
	rm -rf dist build

package:
	goreleaser release --skip=publish --clean

setup:
	git config core.hooksPath .githooks
