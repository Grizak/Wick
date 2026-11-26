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

func newVersionCmd(use string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: "Print detailed version information",
		RunE:  runVersion,
	}
}

func runVersion(cmd *cobra.Command, args []string) error {
	fmt.Printf("%s\n\n", Version())

	// Compatible targets
	fmt.Printf("Compatible targets:\n")
	for target := range target.TargetTriples {
		fmt.Printf("  %s", target)
	}
	fmt.Printf("\nChoose using -t, --target or WICK_TARGET environment variable (default to host)\n\n")

	// LICENSE
	parts := strings.SplitN(assets.License, "\n", 4)
	res := strings.Join([]string{parts[0], parts[2]}, "\n")

	fmt.Printf("%s\n\n", res)

	printChangelog(version)
	return nil
}

func init() {
	rootCmd.AddCommand(newVersionCmd("version"))
	rootCmd.AddCommand(newVersionCmd("v"))
}
