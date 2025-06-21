package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/shortmoose/toobig/internal/base"
	"github.com/shortmoose/toobig/internal/config"
)

var once bool

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
		ref, ix, er := verifyMeta(ctx, path, info)
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
		if update {
			fmt.Printf("Ref '%s': moved to old.\n", path)
		} else {
			fmt.Printf("Ref '%s': to be deleted\n", path)
		}
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

// Validate that all three files (file, ref, and blob) fully match each other.
func verifyMeta(ctx *base.Context, filename string, de fs.DirEntry) (string, error, error) {
	// Verify we have the file ref.
	ref, err := config.ReadFileMeta(filepath.Join(ctx.RefPath, filename))
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("ref doesn't exist"), nil
		}
		return "", nil, fmt.Errorf("reading ref: %w", err)
	}

	// Verify timestamps match.
	info, err := de.Info()
	if err != nil {
		return "", nil, fmt.Errorf("getting file info: %w", err)
	}

	if ref.UnixNano != info.ModTime().UnixNano() {
		return "", fmt.Errorf("file modified"), nil
	}

	// Check if the files are hardlinked.
	info2, err := os.Stat(filepath.Join(ctx.BlobPath, ref.Sha256))
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("blob missing"), nil
		}
		return "", nil, fmt.Errorf("os.Stat of path: %w", err)
	}

	if !os.SameFile(info, info2) {
		return "", fmt.Errorf("file modified"), nil
	}
	return ref.Sha256, nil, nil
}

// Assume "filename" is the source of truth.
func updateMeta(ctx *base.Context, filename string) (string, error) {
	sha256, err := base.GetSha256(filename)
	if err != nil {
		return "", fmt.Errorf("calculating SHA-256: %w", err)
	}

	err = createHardLinkIfNeeded(ctx, filename, sha256)
	if err != nil {
		return "", fmt.Errorf("writing blob: %w", err)
	}

	err = writeFileMeta(ctx, filename, sha256)
	if err != nil {
		return "", fmt.Errorf("writing ref: %w", err)
	}

	return sha256, nil
}

func writeFileMeta(ctx *base.Context, filename, sha256 string) error {
	info, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("stating file: %w", err)
	}

	var fm config.FileMeta
	fm.Sha256 = sha256
	fm.UnixNano = info.ModTime().UnixNano()

	d := filepath.Dir(filepath.Join(ctx.RefPath, filename))
	err = os.MkdirAll(d, 0700)
	if err != nil {
		return fmt.Errorf("making directories: %w", err)
	}

	err = config.WriteFileMeta(ctx.RefPath+"/"+filename, fm)
	if err != nil {
		return fmt.Errorf("writing ref: %w", err)
	}

	return nil
}

func createHardLinkIfNeeded(ctx *base.Context, filename, sha256 string) error {
	// Create a hard link.
	// - If we are already linked is it the correct sha file?
	// - If we aren't already linked, then make it so.
	stat, err := os.Stat(filename)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("file info '%s': %w", filename, err)
	}

	stat2, err := os.Stat(filepath.Join(ctx.BlobPath, sha256))
	if err != nil {
		if os.IsNotExist(err) {
			// Easy case, just create hard link.
			panicIfHashExists(ctx, stat)
			return createHardLink(ctx, filename, sha256)
		}
		return fmt.Errorf("blob info '%s': %w", sha256, err)
	}

	if os.SameFile(stat, stat2) {
		return nil
	}

	panicIfHashExists(ctx, stat)
	// Dup case.
	return createHardLink(ctx, filename, sha256)
}

func createHardLink(ctx *base.Context, filename, sha256 string) error {
	err := os.Link(filename, ctx.BlobPath+"/"+sha256)
	if err == nil {
		fmt.Printf("link created... ")
		return nil
	}

	e, _ := err.(*os.LinkError)
	if e.Err != syscall.EEXIST {
		return fmt.Errorf("link file to blob file: %w", err)
	}
	fmt.Print("dup found...")
	err = mvToOld(ctx, filename, "dup")
	if err != nil {
		return fmt.Errorf("move file to dup directory: %w", err)
	}

	err = os.Link(ctx.BlobPath+"/"+sha256, filename)
	if err != nil {
		return fmt.Errorf("link file to blob file: %w", err)
	}

	fmt.Printf("link created... ")
	return nil
}

func panicIfHashExists(ctx *base.Context, stat os.FileInfo) {
	nlink := uint64(0)
	if sys := stat.Sys(); sys != nil {
		if stat, ok := sys.(*syscall.Stat_t); ok {
			nlink = uint64(stat.Nlink)
		}
	}

	if nlink == 1 {
		// No point walking the whole blob directory.
		return
	}

	err := base.Walk(ctx.BlobPath, func(path string, info fs.DirEntry) error {
		stat2, e := os.Stat(path)
		if e != nil {
			return fmt.Errorf("get inode %s: %w", path, e)
		}

		if os.SameFile(stat, stat2) {
			panic("File corrupted?")
		}
		return nil
	})
	if err != nil {
		panic("Error")
	}
}

func prepareOld(ctx *base.Context) {
	currTime := time.Now()
	path := filepath.Join(ctx.OldPath, currTime.Format("2006-01-02-15:04:05.000"))
	err := os.Mkdir(path, 0755)
	if err != nil {
		panic(err)
	}
	ctx.OldPath = path

	path = filepath.Join(ctx.OldPath, "dup")
	err = os.Mkdir(path, 0755)
	if err != nil {
		panic(err)
	}

	path = filepath.Join(ctx.OldPath, "refs")
	err = os.Mkdir(path, 0755)
	if err != nil {
		panic(err)
	}

	path = filepath.Join(ctx.OldPath, "blobs")
	err = os.Mkdir(path, 0755)
	if err != nil {
		panic(err)
	}
}

func mvToOld(ctx *base.Context, path, sub string) error {
	if !once {
		once = true
		prepareOld(ctx)
	}

	new_path := filepath.Join(ctx.OldPath, sub, strings.ReplaceAll(path, "/", "\\"))
	err := os.Rename(path, new_path)
	if err != nil {
		return fmt.Errorf("move file to dup directory: %w", err)
	}

	return nil
}
