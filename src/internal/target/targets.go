package target

import (
	"fmt"

	"github.com/Grizak/Wick/src/internal/target/targets/darwin"
	"github.com/Grizak/Wick/src/internal/target/targets/linux"
	"github.com/Grizak/Wick/src/internal/target/targets/win"
)

var TargetTriples = map[string]string{
	"linux/amd64":   "x86_64-pc-linux-gnu",
	"linux/arm64":   "aarch64-pc-linux-gnu",
	"darwin/amd64":  "x86_64-apple-darwin",
	"darwin/arm64":  "aarch64-apple-darwin",
	"windows/amd64": "x86_64-pc-windows-msvc",
	"windows/arm64": "aarch64-pc-windows-msvc",
}

type Target interface {
	SysExit() string
	Conv() string
	Triple() string
	EntryPoint() string
	DataLayout() string
}

func NewTarget(target string) (Target, error) {
	switch target {
	case "linux/amd64", "x86_64-pc-linux-gnu":
		return linux.Linux_Amd64{}, nil
	case "linux/arm64", "aarch64-pc-linux-gnu":
		return linux.Linux_Arm64{}, nil
	case "darwin/amd64", "x86_64-apple-darwin":
		return darwin.Darwin_Amd64{}, nil
	case "darwin/arm64", "aarch64-apple-darwin":
		return darwin.Darwin_Arm64{}, nil
	case "windows/amd64", "x86_64-pc-windows-msvc":
		return win.Win_Amd64{}, nil
	case "windows/arm64", "aarch64-pc-windows-msvc":
		return win.Win_Arm64{}, nil
	default:
		return linux.Linux_Amd64{}, fmt.Errorf("unknown target")
	}
}
