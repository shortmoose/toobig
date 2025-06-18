package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	"github.com/shortmoose/toobig/internal/base"
	"github.com/shortmoose/toobig/internal/config"
)

// Restore the files from the combination of a refs and blobs directory.
func Restore(ctx *base.Context) error {
	fmt.Println("Performing restore")

	if !filepath.IsAbs(ctx.FilePathOverride) {
		fmt.Fprintf(os.Stderr, "--file-path=%s isn't a full path.\n", ctx.FilePathOverride)
		os.Exit(3)
	}

	fmt.Println("\nRestoring files:")
	cnt, cnt_e := 0, 0
	err := base.ChdirWalk(ctx.RefPath, func(path string, info fs.DirEntry) error {
		ref, er := config.ReadFileMeta(path)
		if er != nil {
			fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", path, er)
			cnt_e += 1
			return nil
		}

		blob_path := filepath.Join(ctx.BlobPath, ref.Sha256)
		e, er := base.FileExists(blob_path)
		if !e || er != nil {
			fmt.Fprintf(os.Stderr, "Looking for %s: %v\n", blob_path, er)
			cnt_e += 1
			return nil
		}

		files_path := filepath.Join(ctx.FilePathOverride, path)
		d := filepath.Dir(files_path)
		er = os.MkdirAll(d, 0700)
		if er != nil {
			fmt.Fprintf(os.Stderr, "Failed to mkdir %s: %v\n", d, er)
			cnt_e += 1
			return nil
		}

		er = os.Link(blob_path, files_path)
		if er != nil {
			e, _ := er.(*os.LinkError)
			if e.Err != syscall.EEXIST {
				fmt.Fprintf(os.Stderr, "Linking %s to %s: %v\n", files_path, blob_path, e)
				cnt_e += 1
				return nil
			}
			return nil
		}

		cnt += 1
		// Verbose??
		fmt.Printf("%s... LINKED\n", path)

		return nil
	})
	if err != nil {
		return err
	}
	fmt.Printf("%d files restored, %d errors.\n", cnt, cnt_e)

	fmt.Println("\nRestore complete.")
	return nil
}
