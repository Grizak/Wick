package toolchain

import (
	"runtime"
	"testing"
)

func TestLldBinaryNameMatchesHost(t *testing.T) {
	want := "ld.lld"
	switch runtime.GOOS {
	case "windows":
		want = "lld-link.exe"
	case "darwin":
		want = "ld64.lld"
	}
	if got := LldBinaryName(); got != want {
		t.Errorf("LldBinaryName() = %q, want %q on %s", got, want, runtime.GOOS)
	}
}
