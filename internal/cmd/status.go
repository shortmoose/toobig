package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/shortmoose/toobig/internal/base"
)

func Status(ctx *base.Context) error {
	return statusUpdate(ctx, "Status", false)
}

func Update(ctx *base.Context) error {
	return statusUpdate(ctx, "Update", true)
}

func statusUpdate(ctx *base.Context, op string, update bool) error {
	fmt.Printf("Performing %s\n", strings.ToLower(op))

	blob_index := make(map[string]bool)

	fmt.Println("\nExamining files:")
	cnt, cnt_u, cnt_e := 0, 0, 0
	err := base.ChdirWalk(ctx.FilePath, func(path string, info fs.DirEntry) error {
		cnt += 1
		ref, ix, er := verifyMeta(ctx, path)
		if er != nil {
			cnt_e += 1
			fmt.Fprintf(os.Stderr, "File '%s': %v\n", path, er)
			return nil
		}
		if ix == nil {
			blob_index[ref] = true
			return nil
		}

		if update {
			// TODO: Should the old Ref be moved to old??
			// See normal/file_updated test
			// TODO: See update-dup-and-linked, link created multiple times?
			ref, er = updateMeta(ctx, path)
			if er != nil {
				cnt_e += 1
				fmt.Fprintf(os.Stderr, "File '%s': %v\n", path, er)
				return nil
			}
			blob_index[ref] = true
		}

		cnt_u += 1
		fmt.Printf("File '%s': %v\n", path, ix)
		return nil
	})
	if err != nil {
		return err
	}

	if cnt_e != 0 {
		fmt.Fprintf(os.Stderr, "%s failed: %d files, %d updated, %d errors\n", op, cnt, cnt_u, cnt_e)
		os.Exit(11)
	}
	u := (cnt_u > 0)
	fmt.Printf("%d files, %d updated.\n", cnt, cnt_u)

	fmt.Printf("\nValidating refs:\n")
	cnt, cnt_u, cnt_e = 0, 0, 0
	err = base.ChdirWalk(ctx.RefPath, func(path string, info fs.DirEntry) error {
		cnt += 1
		exists, er := base.FileExists(ctx.FilePath + "/" + path)
		if er != nil {
			cnt_e += 1
			fmt.Fprintf(os.Stderr, "Ref '%s': %v\n", path, er)
			return nil
		}
		if exists {
			return nil
		}

		if update {
			er = mvToOld(ctx, path, "refs")
			if er != nil {
				cnt_e += 1
				fmt.Fprintf(os.Stderr, "Ref %s: %v\n", path, er)
				return nil
			}

		}

		cnt_u += 1
		fmt.Printf("Ref '%s': moved to old.\n", path)
		fmt.Printf("Ref '%s': to be deleted\n", path)
		return nil
	})
	if err != nil {
		return err
	}

	if cnt_e != 0 {
		fmt.Fprintf(os.Stderr, "%s failed: %d refs validated, %d errors\n", op, cnt, cnt_e)
		os.Exit(11)
	}
	fmt.Printf("%d refs validated, %d errors.\n", cnt, cnt_e)
	u = (u || cnt_u > 0)

	if update {
		fmt.Printf("\nCleaning up blobs:\n")
		cnt_u, cnt_e = 0, 0
		err = base.ChdirWalk(ctx.BlobPath, func(path string, info fs.DirEntry) error {
			if blob_index[path] {
				return nil
			}

			name := path
			if len(name) == 64 {
				name = name[:8]
			}

			er := mvToOld(ctx, path, "blobs")
			if er != nil {
				fmt.Fprintf(os.Stderr, "Blob '%s': %v\n", name, er)
				cnt_e += 1
				return nil
			}
			cnt_u += 1
			fmt.Printf("Blob '%s': moved to old.\n", name)
			return nil
		})
		if err != nil {
			return err
		}

		if cnt_e != 0 {
			fmt.Fprintf(os.Stderr, "Clean up failed: %d moved, %d errors\n", cnt_u, cnt_e)
			os.Exit(11)
		}
		fmt.Printf("%d moved.\n", cnt_u)
		u = (u || cnt_u > 0)
	}

	fmt.Printf("\n%s complete.\n", op)
	if ctx.UpdateIsError && (cnt_u > 0 || u) {
		os.Exit(10)
	}
	return nil
}
