package cmd

import (
	"fmt"

	"github.com/shortmoose/toobig/internal/base"
	"github.com/shortmoose/toobig/internal/config"
)

// Update the ref and checksum repositories based on the data directory.
func Config(ctx *base.Context) error {
	fmt.Println(`Here is an example config.
Add a path to a directory for each of these settings.
Remember all paths are relative to wherever the config file is located.`)
	fmt.Println("")
	config.WriteExampleConfig()

	return nil
}
