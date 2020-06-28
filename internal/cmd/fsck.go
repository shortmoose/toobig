package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/nthnca/toobig/internal/base"
	"github.com/nthnca/toobig/internal/config"
)

// Do TODO
func Fsck(ctx *base.Context) error {
	fmt.Printf("Performing fsck\n")

	// TODO: Validate the configuration.
	cfg, err := config.ReadConfig(ctx.ConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	ctx.TooBig = cfg

	err = os.Chdir(ctx.DataRepoPath)
	if err != nil {
		log.Fatal(err)
	}

	// Load and validate current set of hashes
	fmt.Printf("Validating blobs:\n")
	err = base.Walk(ctx.HashRepoPath, func(path string, info os.FileInfo) error {
		expected := filepath.Base(path)
		fmt.Printf("%s... validating... ", expected[:8])

		sha, er := base.GetSha256(path)
		if er != nil {
			return er
		}

		if expected != sha {
			fmt.Printf("INVALID: %s\n", sha[:8])
		}

		fmt.Printf("correct\n")
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("\nValidating refs:\n")
	// Walk gitrepo and validate that we have the necessary set of matching hashes.
	err = base.Walk(ctx.GitRepoPath, func(path string, info os.FileInfo) error {
		expected := filepath.Base(path)
		fmt.Printf("Checking: %s\n", expected)

		sha, er := config.ReadFileMeta(path)
		if er != nil {
			return er
		}

		e, er := base.FileExists(filepath.Join(ctx.HashRepoPath, sha.Sha256))
		if !e || er != nil {
			fmt.Printf("BROKE: %v, %v\n", e, er)
		}

		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("Fsck complete.\n")
	return nil
}
