package main

import (
	"fmt"
	"strings"

	"github.com/Grizak/Wick/src/internal/assets"
	"github.com/Grizak/Wick/src/internal/llvm/toolchain"
	"github.com/Grizak/Wick/src/internal/target"
	"github.com/spf13/cobra"
)

var version string

func Version() string {
	return fmt.Sprintf("Wick Version: %s\nLLVM Version: %s", version, toolchain.LLVMVersion())
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print detailed version information",

	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Wick version %s\nLLVM version %s\n\n", version, toolchain.LLVMVersion())

		// Compatible targets
		fmt.Printf("Compatible targets:\n")
		for target := range target.TargetTriples {
			fmt.Printf("  %s", target)
		}
		fmt.Printf("\nChoose using -t, --target or WICK_TARGET environment variable (default to host)\n\n")
		// LICENSE
		fmt.Println(strings.TrimSuffix(assets.License, "\n")) // Remove pending newline
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
