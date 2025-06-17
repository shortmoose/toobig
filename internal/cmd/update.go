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

	fmt.Println("\nUpdating files:")
	cnt, cnt_u, cnt_e := 0, 0, 0
	err := base.ChdirWalk(ctx.FilePath, func(path string, info fs.DirEntry) error {
		cnt += 1
		ix, er := verifyMeta(ctx, path)
		if er != nil {
			cnt_e += 1
			fmt.Fprintf(os.Stderr, "%s: %v", path, er)
			return nil
		}
		if ix == nil {
			return nil
		}

		fmt.Printf("%s: ", path)
		er = updateMeta(ctx, path)
		if er != nil {
			cnt_e += 1
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, er)
			return nil
		}

		cnt_u += 1
		return nil
	})
	if err != nil {
		return err
	}

	if cnt_e != 0 {
		fmt.Fprintf(os.Stderr, "Update failed: %d files, %d updated, %d errors", cnt, cnt_u, cnt_e)
		os.Exit(11)
	}
	u := (cnt_u > 0)
	fmt.Printf("%d files, %d updated.\n", cnt, cnt_u)

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
		er = os.Remove(path)
		if er != nil {
			cnt_e += 1
			fmt.Fprintf(os.Stderr, "Ref:%s: %v\n", path, er)
			return nil
		}

		cnt_u += 1
		return nil
	})
	if err != nil {
		return fmt.Errorf("removing files: %w", err)
	}

	if cnt_e != 0 {
		fmt.Fprintf(os.Stderr, "Update failed: %d files, %d updated, %d errors", cnt, cnt_u, cnt_e)
		os.Exit(11)
	}
	fmt.Printf("%d files, %d updated.\n", cnt, cnt_u)

	fmt.Println("\nUpdate complete.")

	// TODO: How many blobs are no longer needed? Space savings if we delete?
	// TODO: Are there duplicate files?

	if ctx.UpdateIsError && (cnt_u > 0 || u) {
		os.Exit(10)
	}
	return nil
}
