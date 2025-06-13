package cmd

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/shortmoose/toobig/internal/base"
	"github.com/shortmoose/toobig/internal/config"
)

// Fsck TODO
func Fsck(ctx *base.Context) error {
	fmt.Printf("Performing fsck\n")

	// TODO: Validate the configuration.
	cfg, err := config.ReadConfig(ctx.ConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	ctx.TooBig = cfg

	var errors []string
	err = os.Chdir(ctx.DataPath)
	if err != nil {
		log.Fatal(err)
	}

	// Load and validate current set of hashes
	fmt.Printf("Validating blobs:\n")
	err = base.Walk(ctx.HashPath, func(path string, info fs.DirEntry) error {
		expected := filepath.Base(path)
		fmt.Printf("%s... validating... ", expected[:min(len(expected), 8)])

		sha, er := base.GetSha256(path)
		if er != nil {
			return er
		}

		if expected != sha {
			st := fmt.Sprintf("Corrupted blob? %s", sha)
			fmt.Println(st)
			errors = append(errors, st)
			return nil
		}

		fmt.Printf("correct\n")
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("\nValidating refs:\n")
	// Walk gitrepo and validate that we have the necessary set of matching hashes.
	err = base.Walk(ctx.GitRepoPath, func(path string, info fs.DirEntry) error {
		expected := filepath.Base(path)
		fmt.Printf("Checking: %s\n", expected)

		sha, er := config.ReadFileMeta(path)
		if er != nil {
			return er
		}

		e, er := base.FileExists(filepath.Join(ctx.HashPath, sha.Sha256))
		if !e || er != nil {
			st := fmt.Sprintf("No blob stored for %s: %v, %v", path, e, er)
			fmt.Println(st)
			errors = append(errors, st)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if len(errors) != 0 {
		fmt.Printf("Errors: %v\n", errors)
		return fmt.Errorf("bad stuff")
	}
	fmt.Printf("Fsck complete.\n")
	return nil
}
