package main

import (
	"os"

	"github.com/dshelley66/gcp-nuke/cmd"
)

func main() {
	if err := cmd.NewRootCommand().Execute(); err != nil {
		os.Exit(-1)
	}
}
