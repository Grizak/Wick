package main

import (
	"fmt"

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
		fmt.Println(assets.License[:len(assets.License)-1]) // Remove pending newline
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
