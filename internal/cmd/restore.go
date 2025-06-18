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
	// TODO: Should we enforce that this directory is empty?
	if !filepath.IsAbs(ctx.FilePathOverride) {
		fmt.Fprintf(os.Stderr, "--file-path=%s isn't a full path.\n", ctx.FilePathOverride)
		os.Exit(3)
	}

	fmt.Printf("Performing restore to %s\n", ctx.FilePathOverride)

	fmt.Println("\nRestoring files:")
	cnt, cnt_e := 0, 0
	err := base.ChdirWalk(ctx.RefPath, func(path string, info fs.DirEntry) error {
		ref, er := config.ReadFileMeta(path)
		if er != nil {
			fmt.Fprintf(os.Stderr, "%s unreadable: %v\n", path, er)
			cnt_e += 1
			return nil
		}

		blob_path := filepath.Join(ctx.BlobPath, ref.Sha256)
		e, er := base.FileExists(blob_path)
		if !e || er != nil {
			if er != nil {
				fmt.Fprintf(os.Stderr, "%s no blob %s: %v\n", path, ref.Sha256, er)
			} else {
				fmt.Fprintf(os.Stderr, "%s no blob %s\n", path, ref.Sha256)
			}
			cnt_e += 1
			return nil
		}

		files_path := filepath.Join(ctx.FilePathOverride, path)
		d := filepath.Dir(files_path)
		er = os.MkdirAll(d, 0700)
		if er != nil {
			fmt.Fprintf(os.Stderr, "%s failed to mkdir: %v\n", path, er)
			cnt_e += 1
			return nil
		}

		er = os.Link(blob_path, files_path)
		if er != nil {
			e, _ := er.(*os.LinkError)
			if e.Err != syscall.EEXIST {
				fmt.Fprintf(os.Stderr, "%s couldn't link: %v\n", path, e)
				cnt_e += 1
				return nil
			}
			return nil
		}

		cnt += 1
		if ctx.Verbose {
			fmt.Printf("%s... LINKED\n", path)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if cnt_e != 0 {
		fmt.Fprintf(os.Stderr, "Restore failed: %d files restored, %d errors.\n", cnt, cnt_e)
		os.Exit(11)
	}
	fmt.Printf("%d files restored, %d errors.\n", cnt, cnt_e)

	fmt.Println("\nRestore complete.")
	return nil
}
