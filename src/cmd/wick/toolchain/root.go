package toolchain

import "github.com/spf13/cobra"

var ToolchainCmd = &cobra.Command{
	Use:   "toolchain <cmd>",
	Short: "Execute commands with wick's internal toolchain",
	Args:  cobra.MinimumNArgs(1),
}
