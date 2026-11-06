package utils_test

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/Grizak/Wick/src/internal/utils"
)

func TestPrintMdRendersMarkdown(t *testing.T) {
	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = writer
	renderErr := utils.PrintMd("# Wick test\n\nRendered paragraph.")
	closeErr := writer.Close()
	os.Stdout = originalStdout
	if renderErr != nil {
		reader.Close()
		t.Fatalf("PrintMd: %v", renderErr)
	}
	if closeErr != nil {
		reader.Close()
		t.Fatalf("close stdout pipe: %v", closeErr)
	}
	output, err := io.ReadAll(reader)
	reader.Close()
	if err != nil {
		t.Fatalf("read rendered output: %v", err)
	}
	if !strings.Contains(string(output), "Wick test") || !strings.Contains(string(output), "Rendered paragraph.") {
		t.Errorf("rendered output is missing source text: %q", output)
	}
}
