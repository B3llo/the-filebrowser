package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/B3llo/the-filebrowser/sources"
)

func init() {
	sourcesCmd.AddCommand(sourcesAddCmd)
}

var sourcesAddCmd = &cobra.Command{
	Use:   "add <name> <path>",
	Short: "Create a new source",
	Long:  `Create a new source pointing at an absolute path on disk.`,
	Args:  cobra.ExactArgs(2),
	RunE: withStore(func(_ *cobra.Command, args []string, st *store) error {
		path, err := sources.NormalizePath(args[1])
		if err != nil {
			return err
		}

		src := &sources.Source{
			Name:      args[0],
			Path:      path,
			CreatedAt: time.Now(),
		}

		if err := st.Sources.Save(src); err != nil {
			return err
		}
		printSources([]*sources.Source{src})
		return nil
	}, storeOptions{}),
}
