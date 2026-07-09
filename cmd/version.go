package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/B3llo/the-filebrowser/version"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("File Browser v" + version.Version + "/" + version.CommitSHA)
	},
}
