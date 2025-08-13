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
	err := base.ChdirWalk(ctx.FilePath, func(path string, de fs.DirEntry) error {
		cnt += 1

		ref, ix, er := verifyMeta(ctx, path, de)
		if ix == nil && er == nil {
			blob_index[ref] = true
			return nil
		}
		if update {
			ref, er := updateMeta(ctx, path, de)
			if er != nil {
				cnt_e += 1
				fmt.Fprintf(os.Stderr, "File '%s': %v\n", path, er)
				return nil
			}
			blob_index[ref] = true
			cnt_u += 1
			fmt.Printf("File '%s': %v\n", path, "updated")
			return nil
		} else {
			if er != nil {
				cnt_e += 1
				fmt.Fprintf(os.Stderr, "File '%s': %v\n", path, er)
				return nil
			} else {
				cnt_u += 1
				fmt.Printf("File '%s': %v\n", path, ix)
				return nil
			}
		}
	})
	if err != nil {
		return err
	}

	if cnt_e != 0 {
		fmt.Fprintf(os.Stderr, "%s failed: %d files, %d updated, %d errors\n", op, cnt, cnt_u, cnt_e)
		os.Exit(base.ExitCodeDataInconsistency)
	}
	u := (cnt_u > 0)
	fmt.Printf("%d files, %d updated.\n", cnt, cnt_u)

	fmt.Printf("\nValidating refs:\n")
	cnt, cnt_u, cnt_e = 0, 0, 0
	err = base.ChdirWalk(ctx.RefPath, func(path string, info fs.DirEntry) error {
		cnt += 1
		_, er := os.Stat(filepath.Join(ctx.FilePath, path))
		if er == nil {
			return nil
		}
		if !os.IsNotExist(er) {
			cnt_e += 1
			fmt.Fprintf(os.Stderr, "Ref '%s': %v\n", path, er)
			return nil
		}
		// Our ref is out-of-date (os.Stat returned os.IsNotExist).

		if update {
			er = mvToOld(ctx, ctx.RefPath, path, "refs")
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
		os.Exit(base.ExitCodeDataInconsistency)
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

			er := mvToOld(ctx, ctx.BlobPath, path, "blobs")
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
			os.Exit(base.ExitCodeDataInconsistency)
		}
		fmt.Printf("%d moved.\n", cnt_u)
		u = (u || cnt_u > 0)
	}

	fmt.Printf("\n%s complete.\n", op)
	if ctx.UpdateIsError && (cnt_u > 0 || u) {
		os.Exit(base.ExitCodeNormalUpdate)
	}
	return nil
}

// Validate that all three files (file, ref, and blob) fully match each other.
func verifyMeta(ctx *base.Context, filename string, fileEntry fs.DirEntry) (string, error, error) {

	// Load FileInfo for the file.
	info, err := fileEntry.Info()
	if err != nil {
		return "", nil, err
	}

	ref, err := config.ReadFileMeta(filepath.Join(ctx.RefPath, filename))
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("ref doesn't exist"), nil
		}
		return "", nil, fmt.Errorf("reading ref: %w", err)
	}

	info2, err := os.Stat(filepath.Join(ctx.BlobPath, ref.Sha256))
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("blob doesn't exist"), nil
		}
		return "", nil, fmt.Errorf("reading blob: %w", err)
	}

	// Verify timestamps match. Don't return yet.
	time_is_good := ref.UnixNano == info.ModTime().UnixNano()

	if !os.SameFile(info, info2) {
		return "", fmt.Errorf("file modified"), nil
	}

	if !time_is_good {
		return "", nil, fmt.Errorf("file modified, but still hardlinked")
	}

	return ref.Sha256, nil, nil
}

// Treat filename as the source of truth.
func updateMeta(ctx *base.Context, filename string, dEntry fs.DirEntry) (string, error) {
	info, err := dEntry.Info()
	if err != nil {
		return "", err
	}

	sha256, err := base.GetSha256(filename)
	if err != nil {
		return "", fmt.Errorf("calculating SHA-256: %w", err)
	}

	err = createHardLink(ctx, filename, sha256, info)
	if err != nil {
		return "", fmt.Errorf("writing blob: %w", err)
	}

	// The dup case means we need to reload os.Stat.
	info, err = os.Stat(filename)
	if err != nil {
		return "", fmt.Errorf("file stat: %w", err)
	}

	err = mvToOld(ctx, ctx.RefPath, filename, "refs")
	if err != nil {
		return "", fmt.Errorf("ref -> old '%s': %w", filename, err)
	}

	d := filepath.Dir(filepath.Join(ctx.RefPath, filename))
	err = os.MkdirAll(d, 0777) // The umask may adjust these settings.
	if err != nil {
		return "", fmt.Errorf("making directories: %w", err)
	}

	fm := config.FileMeta{Sha256: sha256, UnixNano: info.ModTime().UnixNano()}
	err = config.WriteFileMeta(filepath.Join(ctx.RefPath, filename), fm)
	if err != nil {
		return "", fmt.Errorf("writing ref: %w", err)
	}

	return sha256, nil
}

func createHardLink(ctx *base.Context, filename, sha256 string, stat fs.FileInfo) error {
	blobPath := filepath.Join(ctx.BlobPath, sha256)
	blobInfo, err := os.Stat(blobPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("blob info '%s': %w", sha256, err)
		}

		panicIfHashExists(ctx, stat)

		// Easy case, just create hard link.
		err = os.Link(filename, blobPath)
		if err != nil {
			return err
		}
		return nil
	}

	if os.SameFile(stat, blobInfo) {
		return nil
	}

	panicIfHashExists(ctx, stat)

	// Dup case.
	err = mvToOld(ctx, ctx.FilePath, filename, "dup")
	if err != nil {
		return fmt.Errorf("move file to dup directory: %w", err)
	}

	err = os.Link(blobPath, filename)
	if err != nil {
		return fmt.Errorf("link file to blob file: %w", err)
	}

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
	if once {
		return
	}
	once = true

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

func mvToOld(ctx *base.Context, root, path, sub string) error {
	prepareOld(ctx)

	curr_path := filepath.Join(root, path)
	new_path := filepath.Join(ctx.OldPath, sub, strings.ReplaceAll(path, "/", "\\"))
	err := os.Rename(curr_path, new_path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("move file to dup directory: %w", err)
	}

	return nil
}
