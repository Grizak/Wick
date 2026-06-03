GO = go
BUILD_DIR = build
TARGET ?= wick
TARGET_FULL = $(BUILD_DIR)/$(TARGET)
SOURCES = src/cmd/wick/main.go src/cmd/wick/root.go src/cmd/wick/build.go src/cmd/wick/version.go

.PHONY: all clean art cp

all: cp
	mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags "-X main.version=$(shell git describe --tags --abbrev=0)" -o $(TARGET_FULL) $(SOURCES)

art:
	rm -rf dist 2>/dev/null

clean: art
	rm -rf $(BUILD_DIR) 2>/dev/null

cp: 
	cp LICENSE src/internal/assets/LICENSE
