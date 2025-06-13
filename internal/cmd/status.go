package cmd

import (
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/shortmoose/toobig/internal/base"
	"github.com/shortmoose/toobig/internal/config"
)

func Status(ctx *base.Context) error {
	fmt.Printf("Performing status\n")

	// TODO: Validate the configuration.
	cfg, err := config.ReadConfig(ctx.ConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	ctx.TooBig = cfg

	err = os.Chdir(ctx.FilePath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Scanning data directory...\n")
	cnt := 0
	orphaned := 0
	err = base.Walk(".", func(path string, info fs.DirEntry) error {
		cnt += 1
		valid, er2 := verifyMeta(ctx, path)
		if er2 != nil {
			return fmt.Errorf("verifying metadata: %w", er2)
		}
		if valid {
			return nil
		}
		orphaned += 1
		fmt.Printf("Orphaned data file? %s\n", path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("scanning data directory: %w", err)
	}
	fmt.Printf("%d files checked, %d new or orphaned data files.\n\n", cnt, orphaned)

	err = os.Chdir(ctx.RefPath)
	if err != nil {
		return fmt.Errorf("reading ref directory %s: %w", ctx.RefPath, err)
	}

	fmt.Printf("Scanning ref directory...\n")
	cnt = 0
	orphaned = 0
	// Walk all "meta" files in the git repo.
	err = base.Walk(".", func(path string, info fs.DirEntry) error {
		cnt += 1
		exists, er := base.FileExists(ctx.FilePath + "/" + path)
		if er != nil {
			return fmt.Errorf("check file existence %s: %w", path, er)
		}

		if !exists {
			orphaned += 1
			fmt.Printf("Orphaned metadata file? %s\n", path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("reading git directory: %w", err)
	}
	fmt.Printf("%d files checked, %d new or orphaned data files.\n\n", cnt, orphaned)

	fmt.Printf("Status complete.\n")
	return nil
}
