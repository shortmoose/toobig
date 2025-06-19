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

// Validate that all three files (file, ref, and blob) fully match each other.
func verifyMeta(ctx *base.Context, filename string) (string, error, error) {
	// Verify we have the file ref.
	ref, err := config.ReadFileMeta(filepath.Join(ctx.RefPath, filename))
	if err != nil {
		e, _ := err.(*os.PathError)
		if e.Err == syscall.ENOENT {
			return "", fmt.Errorf("ref doesn't exist"), nil
		}
		return "", nil, fmt.Errorf("reading ref: %w", err)
	}

	// Verify timestamps match.
	info, err := os.Stat(filename)
	if err != nil {
		return "", nil, fmt.Errorf("getting file info: %w", err)
	}

	if ref.UnixNano != info.ModTime().UnixNano() {
		return "", fmt.Errorf("file modified"), nil
	}

	// Verify inodes match.
	inode, err := base.GetInode(filename)
	if err != nil {
		return "", nil, fmt.Errorf("getting inode: %w", err)
	}

	inode2, err := base.GetInode(filepath.Join(ctx.BlobPath, ref.Sha256))
	if err != nil {
		// If the file doesn't exist that isn't really an error.
		e, _ := err.(*os.PathError)
		if e.Err == syscall.ENOENT {
			return "", fmt.Errorf("blob missing"), nil
		}
		return "", nil, fmt.Errorf("getting inode of blob path: %w", err)
	}

	if inode != inode2 {
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
	inode, err := base.GetInode(filename)
	if err != nil {
		return fmt.Errorf("get inode of file: %w", err)
	}

	inode2, err := base.GetInode(ctx.BlobPath + "/" + sha256)
	if err != nil {
		// If the file doesn't exist that isn't really an error.
		e, _ := err.(*os.PathError)
		if e.Err != syscall.ENOENT {
			return fmt.Errorf("get inode of blob path: %w", err)
		}
	}

	// Looks good.
	if inode == inode2 {
		return nil
	}

	currLinkedFile, err := findInodeHash(ctx, inode)
	if err != nil {
		return fmt.Errorf("find inode in hashed files: %w", err)
	}

	if currLinkedFile != "" {
		if sha256 != currLinkedFile {
			return fmt.Errorf("corrupted blob: %s", currLinkedFile)
		}
		return fmt.Errorf("but... we already checked??")
	}

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
	err = lost(ctx, filename, "dup")
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

func findInodeHash(ctx *base.Context, inode uint64) (string, error) {
	var hash string
	err := base.Walk(ctx.BlobPath, func(path string, info fs.DirEntry) error {
		inode2, e := base.GetInode(path)
		if e != nil {
			return fmt.Errorf("get inode %s: %w", path, e)
		}

		if inode == inode2 {
			hash = filepath.Base(path)
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("finding inode: %w", err)
	}

	return hash, nil
}

func foo(ctx *base.Context) {
	currTime := time.Now()

	path := filepath.Join(ctx.OldPath, currTime.Format("2006-01-02-15:04:05.000"))
	fmt.Println(path)
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

func lost(ctx *base.Context, path, sub string) error {

	new_path := filepath.Join(ctx.OldPath, sub, strings.ReplaceAll(path, "/", "-"))
	fmt.Printf("mv %s %s\n", path, new_path)
	err := os.Rename(path, new_path)
	if err != nil {
		return fmt.Errorf("move file to dup directory: %w", err)
	}

	return nil
}
