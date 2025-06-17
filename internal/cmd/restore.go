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

	fmt.Println("\nRestoring files:")
	cnt, cnt_e := 0, 0
	err := base.ChdirWalk(ctx.RefPath, func(path string, info fs.DirEntry) error {
		// Ignore the config file.
		// IGNORE IGNORE IGNORE hi

		ref, er := config.ReadFileMeta(path)
		if er != nil {
			fmt.Fprintf(os.Stderr, "ReadFileMeta failed %s: %v", path, er)
			cnt_e += 1
			return nil
		}

		blob_path := filepath.Join(ctx.BlobPath, ref.Sha256)
		// What is this?
		e, er := base.FileExists(blob_path)
		if !e || er != nil {
			fmt.Fprintf(os.Stderr, "file not found %s: %v", blob_path, er)
			cnt_e += 1
			return nil
		}

		files_path := filepath.Join(ctx.FilePath, path)
		d := filepath.Dir(files_path)
		er = os.MkdirAll(d, 0700)
		if er != nil {
			fmt.Fprintf(os.Stderr, "unable to create directory: %s", files_path)
			cnt_e += 1
			return nil
		}

		er = os.Link(blob_path, files_path)
		if er != nil {
			e, _ := er.(*os.LinkError)
			if e.Err != syscall.EEXIST {
				fmt.Fprintf(os.Stderr, "file not found %s:%v", blob_path, e)
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
