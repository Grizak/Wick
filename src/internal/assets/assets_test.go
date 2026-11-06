package assets_test

import (
	"strings"
	"testing"

	"github.com/Grizak/Wick/src/internal/assets"
)

func TestEmbeddedLicenseAndChangelog(t *testing.T) {
	if strings.TrimSpace(assets.License) == "" {
		t.Fatal("embedded license is empty")
	}

	entries, err := assets.Changelog.ReadDir("changelog")
	if err != nil {
		t.Fatalf("read embedded changelog directory: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("embedded changelog directory is empty")
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			t.Errorf("unexpected changelog entry %q", entry.Name())
		}
		content, err := assets.Changelog.ReadFile("changelog/" + entry.Name())
		if err != nil {
			t.Errorf("read embedded changelog %q: %v", entry.Name(), err)
			continue
		}
		if len(content) == 0 {
			t.Errorf("embedded changelog %q is empty", entry.Name())
		}
	}
}
