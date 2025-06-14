package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/shortmoose/toobig/internal/base"
	"github.com/shortmoose/toobig/internal/config"
)

// Validate that all three files (orig, meta, and hash) fully match each other.
func verifyMeta(ctx *base.Context, filename string) (bool, error) {
	// Verify we have file metadata.
	meta, err := config.ReadFileMeta(filepath.Join(ctx.RefPath, filename))
	if err != nil {
		e, _ := err.(*os.PathError)
		if e.Err == syscall.ENOENT {
			fmt.Printf("meta doesn't exist... ")
			return false, nil
		}
		return false, fmt.Errorf("reading file metadata: %w", err)
	}

	// Verify timestamps match.
	info, err := os.Stat(filename)
	if err != nil {
		return false, fmt.Errorf("stating file: %w", err)
	}

	if meta.UnixNano != info.ModTime().UnixNano() {
		fmt.Printf("file modified... ")
		return false, nil
	}

	// Verify inodes match.
	inode, err := base.GetInode(filename)
	if err != nil {
		return false, fmt.Errorf("getting inode: %w", err)
	}

	inode2, err := base.GetInode(filepath.Join(ctx.BlobPath, meta.Sha256))
	if err != nil {
		// If the file doesn't exist that isn't really an error.
		e, _ := err.(*os.PathError)
		if e.Err == syscall.ENOENT {
			fmt.Printf("link missing... ")
			return false, nil
		}
		return false, fmt.Errorf("getting inode of hash path: %w", err)
	}

	return inode == inode2, nil
}

// Assume "filename" is the source of truth.
func updateMeta(ctx *base.Context, filename string) error {
	// TODO: For some situations (older format repos) we should use
	// meta.Sha256 instead of re-hashing.
	fmt.Printf("hashing... ")
	hash, err := base.GetSha256(filename)
	if err != nil {
		return fmt.Errorf("getting sha256: %w", err)
	}

	err = createHardLinkIfNeeded(ctx, filename, hash)
	if err != nil {
		return fmt.Errorf("creating hard link: %w", err)
	}

	err = writeFileMeta(ctx, filename, hash)
	if err != nil {
		return fmt.Errorf("writing file metadata: %w", err)
	}

	return nil
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
		return fmt.Errorf("writing file metadata: %w", err)
	}

	return nil
}

func createHardLinkIfNeeded(ctx *base.Context, filename, sha256 string) error {
	// Create a hard link.
	// - If we are already linked is it the correct sha file?
	// - If we aren't already linked, then make it so.
	nlink, err := countHardLinks(filename)
	if err != nil {
		return fmt.Errorf("counting hard links: %w", err)
	}

	if nlink == 1 {
		return createHardLink(ctx, filename, sha256)
	}

	inode, err := base.GetInode(filename)
	if err != nil {
		return fmt.Errorf("get inode of data file: %w", err)
	}

	inode2, err := base.GetInode(ctx.BlobPath + "/" + sha256)
	if err != nil {
		// If the file doesn't exist that isn't really an error.
		e, _ := err.(*os.PathError)
		if e.Err != syscall.ENOENT {
			return fmt.Errorf("get inode of hash path: %w", err)
		}
	}

	if inode == inode2 {
		fmt.Printf("link exists... ")
		return nil
	}

	currLinkedFile, err := findInodeHash(ctx, inode)
	if err != nil {
		return fmt.Errorf("find inode in hashed files: %w", err)
	}

	if currLinkedFile != "" {
		if sha256 != currLinkedFile {
			return fmt.Errorf("corrupted hash file: %s", currLinkedFile)
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
		return fmt.Errorf("link file to hash file: %w", err)
	}
	fmt.Print("dup found...")
	fmt.Printf("%s\n", filepath.Join(ctx.DupPath, strings.ReplaceAll(filename, "/", "-")))

	err = os.Rename(filename, filepath.Join(ctx.DupPath, strings.ReplaceAll(filename, "/", "-")))
	if err != nil {
		return fmt.Errorf("move file to dup directory: %w", err)
	}

	err = os.Link(ctx.BlobPath+"/"+sha256, filename)
	if err != nil {
		return fmt.Errorf("link file to hash file: %w", err)
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

func countHardLinks(filename string) (uint64, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return 0, fmt.Errorf("stat file: %w", err)
	}
	nlink := uint64(0)
	if sys := fi.Sys(); sys != nil {
		if stat, ok := sys.(*syscall.Stat_t); ok {
			nlink = uint64(stat.Nlink)
		}
	}

	if nlink == 0 {
		return 0, fmt.Errorf("impossible, how does a file exist with a link count of zero?")
	}

	return nlink, nil
}
