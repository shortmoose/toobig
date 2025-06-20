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

	fmt.Println("\nValidating blobs:")
	curr, cnt, cnt_e := "/", 0, 0
	start := time.Now()
	waiting := true

	err := base.ChdirWalk(ctx.BlobPath, func(path string, info fs.DirEntry) error {
		if len(path) != 64 {
			fmt.Fprintf(os.Stderr, "Blob '%s': doesn't have a valid name\n", path)
			cnt_e += 1
			return nil
		}

		// Display a progress bar (sort of).
		if !ctx.Verbose {
			if curr != path[:len(curr)] {
				curr = path[:len(curr)]
				// TODO: should have a bool that says if we will need a newline.
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
			fmt.Fprintf(os.Stderr, "Blob '%s...': %v\n", path[:8], er)
			cnt_e += 1
			return nil
		}

		if path != sha {
			fmt.Fprintf(os.Stderr, "Blob '%s...': checksum doesn't match\n", path[:8])
			cnt_e += 1
			return nil
		}

		if ctx.Verbose {
			fmt.Printf("Blob '%s...' valid.\n", path[:8])
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

	fmt.Println("\nValidating refs:")
	cnt, cnt_e = 0, 0
	err = base.ChdirWalk(ctx.RefPath, func(path string, info fs.DirEntry) error {

		ref, er := config.ReadFileMeta(path)
		if er != nil {
			fmt.Fprintf(os.Stderr, "'%s': %v\n", path, er)
			cnt_e += 1
			return nil
		}

		blob_path := filepath.Join(ctx.BlobPath, ref.Sha256)
		ex, er := base.FileExists(blob_path)
		if er != nil {
			fmt.Printf("Blob '%s...': %v\n", ref.Sha256[:8], er)
			cnt_e += 1
			return nil
		}
		if !ex {
			fmt.Printf("Blob '%s...' not found.\n", ref.Sha256[:8])
			cnt_e += 1
			return nil
		}

		cnt += 1
		if ctx.Verbose {
			fmt.Printf("Ref '%s': valid.\n", path)
		}
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

	fmt.Println("\nfsck complete.")
	return nil
}
