package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/B3llo/the-filebrowser/sources"
)

func init() {
	rootCmd.AddCommand(sourcesCmd)
}

var sourcesCmd = &cobra.Command{
	Use:   "sources",
	Short: "Sources management utility",
	Long: `Sources management utility.

A source is an independent filesystem location (an absolute path on disk)
that FileBrowser can browse. Users switch between assigned sources from the
sidebar. When no sources are defined, FileBrowser falls back to the classic
per-user scope.`,
	Args: cobra.NoArgs,
}

func printSources(srcs []*sources.Source) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tPath")

	for _, s := range srcs {
		fmt.Fprintf(w, "%d\t%s\t%s\n", s.ID, s.Name, s.Path)
	}

	w.Flush()
}

func parseSourceIDOrName(arg string) (id uint, name string) {
	id64, err := strconv.ParseUint(arg, 10, 64)
	if err != nil {
		return 0, arg
	}
	return uint(id64), ""
}
