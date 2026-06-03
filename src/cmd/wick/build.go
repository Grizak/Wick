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
				KeepOutput:         viper.GetBool("keep-output"),
			},
		)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	flags := buildCmd.Flags()

	flags.StringP("output", "o", "dist/out", "Output file")
	flags.BoolP("save-intermediaries", "s", false, "Save intermediary files")
	flags.StringP("target", "t", "", "Compilation target (default: GOOS/GOARCH, can also be set via WICK_TARGET environment variable)")
	flags.BoolP("keep-output", "k", false, "Don't clear output directory when building")

	viper.BindPFlag("output", flags.Lookup("output"))
	viper.BindPFlag("save-intermediaries", flags.Lookup("save-intermediaries"))
	viper.BindPFlag("target", flags.Lookup("target"))
	viper.BindPFlag("keep-output", flags.Lookup("keep-output"))
}
