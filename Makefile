GO = go
BUILD_DIR = build
TARGET = $(BUILD_DIR)/wickc
SOURCES = src/main.go

.PHONY: all clean

all:
	mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags "-X main.version=$(shell git describe --tags --abbrev=0)" -o $(TARGET) $(SOURCES)

clean:
	rm -rf $(BUILD_DIR) 2>/dev/null
	rm *.ll *.o 2>/dev/null
