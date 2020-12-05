package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/shortmoose/toobig/internal/base"
	"github.com/shortmoose/toobig/internal/config"
)

// Status TODO
func Status(ctx *base.Context) error {
	fmt.Printf("Performing status\n")

	// TODO: Validate the configuration.
	cfg, err := config.ReadConfig(ctx.ConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	ctx.TooBig = cfg

	err = os.Chdir(ctx.DataPath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Scanning data directory...\n")
	err = base.Walk(".", func(path string, info os.FileInfo) error {
		// fmt.Printf("Verifying %s... ", path)
		valid, er2 := verifyMeta(ctx, path)
		if er2 != nil {
			fmt.Printf("\n")
			return fmt.Errorf("verifying metadata: %w", er2)
		}
		if valid {
			// fmt.Printf("Valid")
			return nil
		}
		fmt.Printf("Orphaned data file? %s\n", path)
		return nil
	})

	err = os.Chdir(ctx.GitRepoPath)
	if err != nil {
		return fmt.Errorf("reading data directory %s: %w", ctx.GitRepoPath, err)
	}

	fmt.Printf("Scanning git directory...\n")
	// Walk all "meta" files in the git repo.
	err = base.Walk(".", func(path string, info os.FileInfo) error {
		exists, er := base.FileExists(ctx.DataPath + "/" + path)
		if er != nil {
			return fmt.Errorf("check file existence %s: %w", path, er)
		}

		if !exists {
			fmt.Printf("Orphaned metadata file? %s\n", path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("reading git directory: %w", err)
	}

	fmt.Printf("Status complete.\n")
	return nil
}
