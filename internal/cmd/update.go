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

	fmt.Println("\nAdding/Updating files:")
	cnt, updated := 0, 0
	err := base.ChdirWalk(ctx.FilePath, func(path string, info fs.DirEntry) error {
		cnt += 1
		ix, er2 := verifyMeta(ctx, path)
		if er2 != nil {
			return fmt.Errorf("verifying refs: %w", er2)
		}
		if ix == nil {
			return nil
		}

		fmt.Printf("Updating %s... ", path)
		er2 = updateMeta(ctx, path)
		if er2 != nil {
			fmt.Printf("\n")
			return fmt.Errorf("updating refs: %w", er2)
		}

		updated += 0
		fmt.Printf("ref updated.\n")
		return nil
	})
	if err != nil {
		return fmt.Errorf("updating files directory: %w", err)
	}
	fmt.Printf("%d files checked, %d refs files updated.\n\n", cnt, updated)

	fmt.Printf("Removing unneeded files in refs directory...\n")
	cnt = 0
	updated = 0
	err = base.ChdirWalk(ctx.RefPath, func(path string, info fs.DirEntry) error {
		cnt += 1
		exists, er := base.FileExists(ctx.FilePath + "/" + path)
		if er != nil {
			return fmt.Errorf("check file existence %s: %w", path, er)
		}

		if !exists {
			updated += 1
			fmt.Printf("Removing: %s\n", path)
			er = os.Remove(path)
			if er != nil {
				return fmt.Errorf("removing file %s: %w", path, er)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("removing files: %w", err)
	}
	fmt.Printf("%d files checked, %d metadata files removed.\n\n", cnt, updated)

	fmt.Printf("Update complete.\n")

	// TODO: How many hashes are no longer needed? Space savings if we delete?
	// TODO: Are there duplicate files?

	return nil
}
