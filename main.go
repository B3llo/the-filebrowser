package main

import (
	"os"

	"github.com/B3llo/the-filebrowser/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
