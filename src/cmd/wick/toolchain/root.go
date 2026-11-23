package toolchain

import "github.com/spf13/cobra"

var ToolchainCmd = &cobra.Command{
	Use:  "toolchain <cmd>",
	Args: cobra.MinimumNArgs(1),
}
