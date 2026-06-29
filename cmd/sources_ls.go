package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	sourcesCmd.AddCommand(sourcesLsCmd)
}

var sourcesLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List existing sources",
	Args:  cobra.NoArgs,
	RunE: withStore(func(_ *cobra.Command, _ []string, st *store) error {
		srcs, err := st.Sources.Gets()
		if err != nil {
			return err
		}
		printSources(srcs)
		return nil
	}, storeOptions{}),
}
