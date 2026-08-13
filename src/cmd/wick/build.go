package main

import (
	"fmt"
	"runtime"

	"github.com/Grizak/Wick/src/internal/compiler"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var buildCmd = &cobra.Command{
	Use:   "build [options] <source files>",
	Short: "Build Wick source files",
	Args:  cobra.MinimumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {

		target := viper.GetString("target")

		if target == "" {
			target = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
		}

		return compiler.Build(
			compiler.BuildOptions{
				Input:              args,
				Output:             viper.GetString("output"),
				SaveIntermediaries: viper.GetBool("save-intermediaries"),
				Target:             target,
				Opt:                viper.GetInt("optimization"),
			},
		)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
