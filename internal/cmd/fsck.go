package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/shortmoose/toobig/internal/base"
	"github.com/shortmoose/toobig/internal/config"
)

// Basic details:
// fsck, minimal output, display errors.
// for each section it should output the total number of files it processed.
func Fsck(ctx *base.Context) error {
	fmt.Println("Performing fsck")

	err := os.Chdir(ctx.FilePath)
	if err != nil {
		return fmt.Errorf("cd %s: %w", ctx.FilePath, err)
	}

	var errors []string

	// Load and validate current set of hashes
	fmt.Println("Validating blobs:")
	a := "s"
	c := 0
	e := 0
	err = base.Walk(ctx.BlobPath, func(path string, info fs.DirEntry) error {
		expected := filepath.Base(path)
		if a != expected[:1] {
			a = expected[:1]
			fmt.Printf("%s...", a)
		}

		if ctx.Verbose {
			fmt.Printf("%s... validating... ", expected[:min(len(expected), 8)])
		}

		sha, er := base.GetSha256(path)
		if er != nil {
			return er
		}

		if expected != sha {
			st := fmt.Sprintf("Corrupted blob? %s", sha)
			fmt.Println(st)
			errors = append(errors, st)
			e += 1
			return nil
		}

		if ctx.Verbose {
			fmt.Printf("correct\n")
		}
		c += 1
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Printf("\n%d blobs validated, %d errors.\n", c, e)

	c = 0
	e = 0
	fmt.Printf("\nValidating refs:\n")
	// Walk gitrepo and validate that we have the necessary set of matching hashes.
	err = base.Walk(ctx.RefPath, func(path string, info fs.DirEntry) error {
		expected := filepath.Base(path)
		if ctx.Verbose {
			fmt.Printf("Checking: %s\n", expected)
		}

		sha, er := config.ReadFileMeta(path)
		if er != nil {
			return er
		}

		ex, er := base.FileExists(filepath.Join(ctx.BlobPath, sha.Sha256))
		if !ex || er != nil {
			st := fmt.Sprintf("No blob stored for %s: %v, %v", path, ex, er)
			fmt.Println(st)
			errors = append(errors, st)
			e += 1
			return nil
		}

		c += 1
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Printf("%d refs validated, %d errors.\n", c, e)

	if len(errors) != 0 {
		fmt.Printf("Errors: %v\n", errors)
		return fmt.Errorf("bad stuff")
	}
	fmt.Println("\nFsck complete.")
	return nil
}
