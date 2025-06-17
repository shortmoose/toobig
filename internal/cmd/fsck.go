package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/shortmoose/toobig/internal/base"
	"github.com/shortmoose/toobig/internal/config"
)

// Basic details:
// fsck, minimal output, display errors.
// for each section it should output the total number of files it processed.
func Fsck(ctx *base.Context) error {
	fmt.Println("Performing fsck")

	// ########
	fmt.Println("\nValidating blobs:")
	curr, cnt, cnt_e := "/", 0, 0
	start := time.Now()
	waiting := true

	err := base.ChdirWalk(ctx.BlobPath, func(path string, info fs.DirEntry) error {
		filename := filepath.Base(path)

		// Display a progress bar (sort of).
		if !ctx.Verbose {
			if len(filename) > len(curr) && curr != filename[:len(curr)] {
				curr = filename[:len(curr)]
				fmt.Printf("%s..", curr)
			}

			// Increase granularity if this is going to take a long time.
			if waiting && time.Since(start).Seconds() > 100 {
				waiting = false
				if curr == "0" {
					curr = "//"
				}
			}
		}

		sha, er := base.GetSha256(path)
		if er != nil {
			fmt.Fprintf(os.Stderr, "Blob '%s' failed: %v\n", filename, er)
			cnt_e += 1
			return nil
		}

		if filename != sha {
			fmt.Fprintf(os.Stderr, "Blob '%s' appears corrupted: %s\n", filename, sha)
			cnt_e += 1
			return nil
		}

		if ctx.Verbose {
			fmt.Printf("Blob '%s...' valid.\n", filename[:8])
		}

		cnt += 1
		return nil
	})
	if err != nil {
		return err
	}

	if !ctx.Verbose {
		fmt.Println("")
	}
	if cnt_e != 0 {
		fmt.Fprintf(os.Stderr, "\nFsck failed: %d blobs validated, %d errors\n", cnt, cnt_e)
		os.Exit(11)
	}
	fmt.Printf("%d blobs validated, %d errors.\n", cnt, cnt_e)

	// ########
	fmt.Println("\nValidating refs:")
	cnt, cnt_e = 0, 0
	err = base.ChdirWalk(ctx.RefPath, func(path string, info fs.DirEntry) error {
		filename := filepath.Base(path)

		sha, er := config.ReadFileMeta(path)
		if er != nil {
			fmt.Printf("Ref '%s' invalid: %v\n", path, er)
			cnt_e += 1
			return nil
		}

		ex, er := base.FileExists(filepath.Join(ctx.BlobPath, sha.Sha256))
		if !ex || er != nil {
			fmt.Printf("Ref '%s' doesn't point to a blob: %v\n", path, er)
			cnt_e += 1
			return nil
		}

		if ctx.Verbose {
			fmt.Printf("Ref '%s' valid.\n", filename)
		}

		cnt += 1
		return nil
	})
	if err != nil {
		return err
	}

	if cnt_e != 0 {
		fmt.Fprintf(os.Stderr, "\nFsck failed: %d refs validated, %d errors\n", cnt, cnt_e)
		os.Exit(11)
	}
	fmt.Printf("%d refs validated, %d errors.\n", cnt, cnt_e)

	fmt.Println("\nFsck complete.")
	return nil
}
