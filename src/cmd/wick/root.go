package main

import (
	"runtime"

	"github.com/Grizak/Wick/src/internal/compiler"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "wick [options/subcommand] <source files>",
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

	flags := rootCmd.PersistentFlags()

	flags.StringP("output", "o", "dist/out", "Output file")
	flags.BoolP("save-intermediaries", "s", false, "Save intermediary files")
	flags.StringP("target", "t", "", "Compilation target (default: same as host, can also be set via WICK_TARGET environment variable)")

	viper.BindPFlag("output", flags.Lookup("output"))
	viper.BindPFlag("save-intermediaries", flags.Lookup("save-intermediaries"))
	viper.BindPFlag("target", flags.Lookup("target"))

	rootCmd.SetVersionTemplate("{{.Version}}\n")
}
