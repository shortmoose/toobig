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

func Fsck(ctx *base.Context) error {
	fmt.Println("Performing fsck")

	fmt.Println("\nValidating blobs:")
	cnt, cnt_e, curr, start, waiting := 0, 0, "/", time.Now(), true

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

	return fsckRestore(ctx, "Fsck", false)
}

func Restore(ctx *base.Context) error {
	// TODO: Should we enforce that this directory is empty?
	if !filepath.IsAbs(ctx.FilePathOverride) {
		fmt.Fprintf(os.Stderr, "--file-path=%s isn't a full path.\n", ctx.FilePathOverride)
		os.Exit(3)
	}

	fmt.Printf("Performing restore to %s\n", ctx.FilePathOverride)

	return fsckRestore(ctx, "Restore", true)
}

func fsckRestore(ctx *base.Context, op string, restore bool) error {
	fmt.Println("\nValidating refs:")

	cnt, cnt_e := 0, 0
	err := base.ChdirWalk(ctx.RefPath, func(path string, info fs.DirEntry) error {
		ref, er := config.ReadFileMeta(path)
		if er != nil {
			fmt.Fprintf(os.Stderr, "Ref '%s': %v\n", path, er)
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

		if restore {
			files_path := filepath.Join(ctx.FilePathOverride, path)
			d := filepath.Dir(files_path)
			er = os.MkdirAll(d, 0700)
			if er != nil {
				fmt.Fprintf(os.Stderr, "mkdir '%s': %v\n", d, er)
				cnt_e += 1
				return nil
			}

			er = os.Link(blob_path, files_path)
			if er != nil {
				fmt.Fprintf(os.Stderr, "linking %s to %s: %v\n", files_path, blob_path, er)
				cnt_e += 1
				return nil
			}
		}

		cnt += 1
		if ctx.Verbose {
			if restore {
			fmt.Printf("Ref '%s': restored.\n", path)
			} else {
			fmt.Printf("Ref '%s': valid.\n", path)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if cnt_e != 0 {
		fmt.Fprintf(os.Stderr, "\n%s failed: %d refs validated, %d errors\n", op, cnt, cnt_e)
		os.Exit(11)
	}
	fmt.Printf("%d refs validated, %d errors.\n", cnt, cnt_e)

	fmt.Printf("\n%s complete.\n", op)
	return nil
}
