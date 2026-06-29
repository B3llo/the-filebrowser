package cmd

import (
	"github.com/spf13/cobra"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
)

func init() {
	sourcesCmd.AddCommand(sourcesRmCmd)
}

var sourcesRmCmd = &cobra.Command{
	Use:   "rm <id|name>",
	Short: "Delete an existing source",
	Args:  cobra.ExactArgs(1),
	RunE: withStore(func(_ *cobra.Command, args []string, st *store) error {
		id, name := parseSourceIDOrName(args[0])
		if name != "" {
			src, err := st.Sources.Get(name)
			if err != nil {
				return err
			}
			id = src.ID
		}
		if id == 0 {
			return fberrors.ErrNotExist
		}
		return st.Sources.Delete(id)
	}, storeOptions{}),
}
