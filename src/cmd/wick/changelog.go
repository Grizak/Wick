package main

import (
	"fmt"

	"github.com/Grizak/Wick/src/internal/assets"
	"github.com/Grizak/Wick/src/internal/utils"
	"github.com/spf13/cobra"
)

var changelogCmd = &cobra.Command{
	Use:   "changelog [version]",
	Short: "Print the changelog of Wick",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v := version
		if len(args) == 1 {
			v = args[0]
		}

		cmd.SilenceErrors = true
		cmd.SilenceUsage = true

		fmt.Printf("%s:\n", v)

		return printChangelog(v)
	},
}

func init() {
	rootCmd.AddCommand(changelogCmd)
}

func printChangelog(v string) error {
	changelog, err := assets.Changelog.ReadFile(fmt.Sprintf("changelog/%s.md", v))
	if err != nil {
		return fmt.Errorf("read changelog for %q: %w", v, err)
	}

	utils.PrintMd(string(changelog))
	return nil
}
