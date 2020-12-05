package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/shortmoose/toobig/internal/base"
	"github.com/shortmoose/toobig/internal/config"
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

	// Walk gitrepo and validate that we have the necessary set of
	// matching hashes.
	fmt.Printf("Restoring files:\n")
	err = base.Walk(".", func(path string, info os.FileInfo) error {
		// Ignore the config file.
		fmt.Printf("%s... ", path)

		sha, er := config.ReadFileMeta(path)
		if er != nil {
			return fmt.Errorf("ReadFileMeta failed %s: %w", path, er)
		}

		hashFile := filepath.Join(ctx.HashPath, sha.Sha256)

		e, er := base.FileExists(hashFile)
		if !e || er != nil {
			return fmt.Errorf("file not found %s: %w", hashFile, er)
		}

		dataPath := filepath.Join(ctx.DataPath, path)
		d := filepath.Dir(dataPath)
		er = os.MkdirAll(d, 0700)
		if er != nil {
			return nil
		}

		er = os.Link(hashFile, dataPath)
		if er != nil {
			e, _ := er.(*os.LinkError)
			if e.Err != syscall.EEXIST {
				return fmt.Errorf("file not found %s:%w", hashFile, e)
			}
			return nil
		}
		fmt.Printf("LINKED\n")

		return nil
	})
	if err != nil {
		fmt.Printf("\n")
		return err
	}

	return nil
}
