package main

import (
	"strings"

	"github.com/Grizak/Wick/src/internal/llvm/linker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var linkCmd = &cobra.Command{
	Use:   "link <triple> --ldflags=\"<link flags>\"",
	Short: "Execute wick's internal linker",
	Args:  cobra.MinimumNArgs(1),
	RunE:  link,
}

func link(cmd *cobra.Command, args []string) error {
	triple := args[0]
	return linker.ExternalLink(triple, strings.Split(viper.GetString("ldflags"), " "))
}

func init() {
	rootCmd.AddCommand(linkCmd)
}
