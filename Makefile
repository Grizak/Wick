GO = go
BUILD_DIR = build
TARGET ?= wick
TARGET_FULL = $(BUILD_DIR)/$(TARGET)
SOURCES = src/cmd/wick/main.go src/cmd/wick/root.go src/cmd/wick/build.go src/cmd/wick/version.go

.PHONY: current win linux darwin all clean art cp

current: prepare
	./build.sh $(TARGET_FULL) $(SOURCES)

win: prepare
	GOOS=windows GOARCH=amd64 ./build.sh $(TARGET_FULL) $(SOURCES)

linux: prepare
	GOOS=linux GOARCH=amd64 ./build.sh $(TARGET_FULL) $(SOURCES)

darwin: prepare
	GOOS=darwin GOARCH=arm64 ./build.sh $(TARGET_FULL) $(SOURCES)

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
