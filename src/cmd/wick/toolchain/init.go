package toolchain

import (
	tools "github.com/Grizak/Wick/src/internal/llvm/toolchain"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the wick toolchain onto disk",
	RunE: func(_ *cobra.Command, _ []string) error {
		return tools.Init()
	},
}

func init() {
	ToolchainCmd.AddCommand(initCmd)
}
