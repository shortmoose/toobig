package cmd

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/shortmoose/toobig/internal/base"
)

func Status(ctx *base.Context) error {
	fmt.Println("Performing status")

	// ########
	fmt.Println("\nValidating files:")
	cnt, cnt_u, cnt_e := 0, 0, 0
	err := base.ChdirWalk(ctx.FilePath, func(path string, info fs.DirEntry) error {
		cnt += 1
		ix, er := verifyMeta(ctx, path)
		if er != nil {
			cnt_e += 1
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, er)
			return nil
		}
		if ix == nil {
			return nil
		}

		cnt_u += 1
		fmt.Printf("%s: %v\n", path, ix)
		return nil
	})
	if err != nil {
		return err
	}

	if cnt_e != 0 {
		fmt.Fprintf(os.Stderr, "Update failed: %d files, %d updated, %d errors", cnt, cnt_u, cnt_e)
		os.Exit(14)
	}
	u := (cnt_u > 0)
	fmt.Printf("%d files, %d updated.\n", cnt, cnt_u)

	// ########
	fmt.Printf("\nValidating refs:\n")
	cnt, cnt_u, cnt_e = 0, 0, 0
	err = base.ChdirWalk(ctx.RefPath, func(path string, info fs.DirEntry) error {
		cnt += 1
		exists, er := base.FileExists(ctx.FilePath + "/" + path)
		if er != nil {
			cnt_e += 1
			fmt.Fprintf(os.Stderr, "Ref '%s' err: %v\n", path, er)
			return nil
		}
		if exists {
			return nil
		}

		cnt_u += 1
		fmt.Printf("Ref:%s to be deleted.\n", path)
		return nil
	})
	if err != nil {
		return err
	}

	if cnt_e != 0 {
		fmt.Fprintf(os.Stderr, "%d refs validated, %d errors", cnt, cnt_e)
		os.Exit(14)
	}
	fmt.Printf("%d refs validated, %d errors.\n", cnt, cnt_e)

	fmt.Println("\nStatus complete.")
	if ctx.Verbose && (cnt_u > 0 || u) {
		os.Exit(13)
	}
	return nil
}
