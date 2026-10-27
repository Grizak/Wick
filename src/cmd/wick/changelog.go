package main

import (
	"fmt"

	"github.com/Grizak/Wick/src/internal/assets"
	"github.com/Grizak/Wick/src/internal/utils"
	"github.com/spf13/cobra"
)

var changelogCmd = &cobra.Command{
	Use:   "changelog <version>",
	Short: "Print the changelog of Wick",
	Run: func(cmd *cobra.Command, args []string) {
		var v string
		if len(args) < 1 {
			v = version
		} else {
			v = args[0]
		}
		changelog, err := assets.Changelog.ReadFile(fmt.Sprintf("changelog/%s.md", v))
		if err != nil {
			fmt.Println("Error reading changelog:", err)
			return
		}
		fmt.Printf("%s:\n", v)
		utils.PrintMd(string(changelog))
	},
}

func init() {
	rootCmd.AddCommand(changelogCmd)
}
