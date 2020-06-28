package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nthnca/toobig/internal/base"
	"github.com/nthnca/toobig/internal/config"
)

// Restore the data from the combination of a ref and checksum repository.
func Restore(ctx *base.Context) error {
	fmt.Printf("Performing restore:\n")

	// TODO: Validate the configuration.
	cfg, err := config.ReadConfig(ctx.ConfigPath)
	if err != nil {
		return err
	}
	ctx.TooBig = cfg

	// Set our current working directory to the git path.
	err = os.Chdir(ctx.GitRepoPath)
	if err != nil {
		return err
	}

	// Walk gitrepo and validate that we have the necessary set of matching hashes.
	fmt.Printf("Restoring files:\n")
	err = base.Walk(".", func(path string, info os.FileInfo) error {
		// Ignore the config file.
		fmt.Printf("%s... ", path)

		sha, er := config.ReadFileMeta(path)
		if er != nil {
			return er
		}

		hashFile := filepath.Join(ctx.HashRepoPath, sha.Sha256)

		e, er := base.FileExists(hashFile)
		if !e || er != nil {
			fmt.Printf("BROKE: %v, %v\n", e, er)
			return nil
		}

		dataPath := filepath.Join(ctx.DataRepoPath, path)
		d := filepath.Dir(dataPath)
		er = os.MkdirAll(d, 0700)
		if er != nil {
			return nil
		}

		er = os.Link(hashFile, dataPath)
		if er != nil {
			return nil
		}
		fmt.Printf("LINKED\n")

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
