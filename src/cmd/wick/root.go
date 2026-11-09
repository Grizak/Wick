package main

import (
	"strings"

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

	RunE: runBuild,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	viper.SetEnvPrefix("WICK")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	flags := rootCmd.PersistentFlags()

	flags.StringP("output", "o", "dist/out", "Output file")
	flags.BoolP("save-intermediaries", "s", false, "Save intermediary files")
	flags.StringP("target", "t", "", "Compilation target (default: same as host, can also be set via WICK_TARGET environment variable)")
	flags.IntP("optimization", "O", 0, "Optimization level (0-3)")

	viper.BindPFlag("output", flags.Lookup("output"))
	viper.BindPFlag("save-intermediaries", flags.Lookup("save-intermediaries"))
	viper.BindPFlag("target", flags.Lookup("target"))
	viper.BindPFlag("optimization", flags.Lookup("optimization"))

	rootCmd.SetVersionTemplate("{{.Version}}\n")
}
