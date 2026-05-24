package backends

import (
	"fmt"

	"github.com/Grizak/Wick/src/backends/linux_amd64"
	"github.com/Grizak/Wick/src/types"
)

func New(target string) (types.Backend, error) {
	switch target {
	case "linux/amd64":
		return linux_amd64.NewLinuxAMD64Backend(), nil
	default:
		return nil, fmt.Errorf("unsupported target: %s", target)
	}
}
