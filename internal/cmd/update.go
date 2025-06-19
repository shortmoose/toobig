package cmd

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/shortmoose/toobig/internal/base"
)

// Update the refs and blobs directories based on the files directory.
func Update(ctx *base.Context) error {
	fmt.Println("Performing update")

	foo(ctx)
	blob_index := make(map[string]bool)

	fmt.Println("\nUpdating files:")
	cnt, cnt_u, cnt_e := 0, 0, 0
	err := base.ChdirWalk(ctx.FilePath, func(path string, info fs.DirEntry) error {
		cnt += 1
		ref, ix, er := verifyMeta(ctx, path)
		if er != nil {
			cnt_e += 1
			fmt.Fprintf(os.Stderr, "%s: %v", path, er)
			return nil
		}
		if ix == nil {
			blob_index[ref] = true
			return nil
		}

		ref, er = updateMeta(ctx, path)
		if er != nil {
			cnt_e += 1
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, er)
			return nil
		}
		blob_index[ref] = true

		cnt_u += 1
		return nil
	})
	if err != nil {
		return err
	}

	if cnt_e != 0 {
		fmt.Fprintf(os.Stderr, "Update failed: %d files, %d updated, %d errors\n", cnt, cnt_u, cnt_e)
		os.Exit(11)
	}
	u := (cnt_u > 0)
	fmt.Printf("%d files, %d updated.\n", cnt, cnt_u)

	// Grab all the checksum numbers.

	fmt.Printf("\nCleaning up refs:\n")
	cnt, cnt_u, cnt_e = 0, 0, 0
	err = base.ChdirWalk(ctx.RefPath, func(path string, info fs.DirEntry) error {
		cnt += 1
		exists, er := base.FileExists(ctx.FilePath + "/" + path)
		if er != nil {
			cnt_e += 1
			fmt.Fprintf(os.Stderr, "Ref:%s: %v\n", path, er)
			return nil
		}
		if exists {
			return nil
		}

		fmt.Printf("Ref:%s deleted\n", path)
		er = lost(ctx, path, "refs")
		if er != nil {
			cnt_e += 1
			fmt.Fprintf(os.Stderr, "Ref:%s: %v\n", path, er)
			return nil
		}

		cnt_u += 1
		return nil
	})
	if err != nil {
		return err
	}

	if cnt_e != 0 {
		fmt.Fprintf(os.Stderr, "Update failed: %d files, %d updated, %d errors\n", cnt, cnt_u, cnt_e)
		os.Exit(11)
	}
	fmt.Printf("%d files, %d updated.\n", cnt, cnt_u)

	// Create an index of all checksums
	err = base.ChdirWalk(ctx.BlobPath, func(path string, info fs.DirEntry) error {

		if blob_index[path] {
			return nil
		}

		er := lost(ctx, path, "blobs")
		if er != nil {
			fmt.Fprintf(os.Stderr, "Ref:%s: %v\n", path, er)
			return nil
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Println("\nUpdate complete.")

	if ctx.UpdateIsError && (cnt_u > 0 || u) {
		os.Exit(10)
	}
	return nil
}
