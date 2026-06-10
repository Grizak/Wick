GO = go
BUILD_DIR ?= build
TARGET ?= wick
DARWIN_TARGET ?= $(TARGET)
WIN_TARGET ?= $(TARGET)
LINUX_TARGET ?= $(TARGET)
TARGET_FULL = $(BUILD_DIR)/$(TARGET)
DARWIN_TARGET_FULL = $(BUILD_DIR)/$(DARWIN_TARGET)
WIN_TARGET_FULL = $(BUILD_DIR)/$(WIN_TARGET)
LINUX_TARGET_FULL = $(BUILD_DIR)/$(LINUX_TARGET)
SOURCES = src/cmd/wick/main.go src/cmd/wick/root.go src/cmd/wick/build.go src/cmd/wick/version.go

.PHONY: current win linux darwin all clean art cp

current: prepare
	./build.sh $(TARGET_FULL) $(SOURCES)

win: prepare
	GOOS=windows GOARCH=amd64 ./build.sh $(WIN_TARGET_FULL) $(SOURCES)

linux: prepare
	GOOS=linux GOARCH=amd64 ./build.sh $(LINUX_TARGET_FULL) $(SOURCES)

darwin: prepare
	GOOS=darwin GOARCH=arm64 ./build.sh $(DARWIN_TARGET_FULL) $(SOURCES)

all: win linux darwin
	@echo "Done"

prepare: assets
	mkdir -p $(BUILD_DIR)

art:
	rm -rf dist 2>/dev/null

clean: art
	rm -rf $(BUILD_DIR) 2>/dev/null

assets: 
	cp LICENSE src/internal/assets/LICENSE
