package main

import (
	"runtime"

	"github.com/Grizak/Wick/src/internal/compiler"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "wick",
	Short: "The Wick compiler",

	// --version
	Version: version,

	// Make "wick build" and "wick" do the same thing (i.e. "build" is the default command)
	// for backwards compatibility with older versions of the compiler and convenience
	Args: cobra.MinimumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		target := viper.GetString("target")

		if target == "" {
			target = runtime.GOOS + "/" + runtime.GOARCH
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

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	viper.SetEnvPrefix("WICK")
	viper.AutomaticEnv()

	rootCmd.SetVersionTemplate("{{.Version}}\n")
}
