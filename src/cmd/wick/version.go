package main

import (
	"fmt"
	"strings"

	"github.com/Grizak/Wick/src/internal/assets"
	"github.com/spf13/cobra"
)

var version string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print detailed version information",

	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Wick version %s\n\n", version)
		// LICENSE
		fmt.Println(strings.TrimSuffix(assets.License, "\n")) // Remove pending newline
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
