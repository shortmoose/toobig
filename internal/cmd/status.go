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
	cnt := 0
	cnt_e := 0
	err := base.ChdirWalk(ctx.FilePath, func(path string, info fs.DirEntry) error {
		valid, er := verifyMeta(ctx, path)
		if er != nil {
			fmt.Fprintf(os.Stderr, "File %s ref status: %v\n", path, er)
			cnt_e += 1
			return nil
		}

		if !valid {
			fmt.Fprintf(os.Stderr, "File %s missing ref file\n", path)
			cnt_e += 1
			return nil
		}

		if ctx.Verbose {
			fmt.Printf("File %s has matching ref file.\n", path)
		}

		cnt += 1
		return nil
	})
	if err != nil {
		return err
	}

	if cnt_e != 0 {
		return fmt.Errorf("%d files validated, %d errors", cnt, cnt_e)
	}
	fmt.Printf("%d files validated, %d errors.\n", cnt, cnt_e)

	// ########
	fmt.Println("\nValidating refs:")
	cnt = 0
	cnt_e = 0
	err = base.ChdirWalk(ctx.RefPath, func(path string, info fs.DirEntry) error {
		exists, er := base.FileExists(ctx.FilePath + "/" + path)
		if er != nil {
			fmt.Fprintf(os.Stderr, "Ref %s err: %v\n", path, er)
			cnt_e += 1
			return nil
		}

		if !exists {
			fmt.Fprintf(os.Stderr, "Stale ref %s (will be deleted on update)\n", path)
			cnt_e += 1
			return nil
		}

		if ctx.Verbose {
			fmt.Printf("Ref %s matches file.\n", path)
		}

		cnt += 1
		return nil
	})
	if err != nil {
		return err
	}

	if cnt_e != 0 {
		return fmt.Errorf("%d refs validated, %d errors", cnt, cnt_e)
	}
	fmt.Printf("%d refs validated, %d errors.\n", cnt, cnt_e)

	fmt.Printf("Status complete.\n")
	return nil
}
