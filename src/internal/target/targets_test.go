package target_test

import (
	"testing"

	"github.com/Grizak/Wick/src/internal/target"
)

func TestNewTargetForAliases(t *testing.T) {
	for conventional, triple := range target.TargetTriples {
		t.Run(conventional, func(t *testing.T) {
			byName, err := target.NewTarget(conventional)
			if err != nil {
				t.Fatalf("create target from name: %v", err)
			}
			if got := byName.Conv(); got != conventional {
				t.Errorf("convention = %q, want %q", got, conventional)
			}
			if got := byName.Triple(); got != triple {
				t.Errorf("triple = %q, want %q", got, triple)
			}
			if byName.EntryPoint() == "" || byName.DataLayout() == "" || byName.SysExit() == "" {
				t.Error("target is missing required code-generation properties")
			}

			byTriple, err := target.NewTarget(triple)
			if err != nil {
				t.Fatalf("create target from triple: %v", err)
			}
			if byTriple.Conv() != conventional {
				t.Errorf("triple alias resolves to %q, want %q", byTriple.Conv(), conventional)
			}
		})
	}
}

func TestNewTargetRejectsUnknownTarget(t *testing.T) {
	if _, err := target.NewTarget("freebsd/amd64"); err == nil {
		t.Fatal("unknown target accepted")
	}
}
