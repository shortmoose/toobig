package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/shortmoose/toobig/internal/base"
	"github.com/shortmoose/toobig/internal/config"
)

// Update the ref and checksum repositories based on the data directory.
func Update(ctx *base.Context) error {
	fmt.Printf("Performing update:\n")

	cfg, err := config.ReadConfig(ctx.ConfigPath)
	if err != nil {
		return err
	}
	ctx.TooBig = cfg

	err = os.Chdir(ctx.DataPath)
	if err != nil {
		return err
	}

	fmt.Printf("Updating  data directory...\n")
	err = base.Walk(".", func(path string, info os.FileInfo) error {
		fmt.Printf("Verifying %s... ", path)
		valid, er2 := verifyMeta(ctx, path)
		if er2 != nil {
			fmt.Printf("\n")
			return er2
		}
		if valid {
			fmt.Printf("Valid\n")
			return nil
		}

		fmt.Printf("stale... ")
		er2 = updateMeta(ctx, path)
		if er2 != nil {
			fmt.Printf("\n")
			return er2
		}

		fmt.Printf("metadata updated.\n")
		return nil
	})
	if err != nil {
		return err
	}

	err = os.Chdir(ctx.GitRepoPath)
	if err != nil {
		return err
	}

	fmt.Printf("Removing uneeded files in git directory...\n")
	// Walk all "meta" files in the git repo.
	err = base.Walk(".", func(path string, info os.FileInfo) error {
		exists, er := base.FileExists(ctx.DataPath + "/" + path)
		if er != nil {
			return er
		}

		if !exists {
			fmt.Printf("Removing: %s\n", path)
			e := os.Remove(path)
			if e != nil {
				return e
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("Update complete.\n")
	// TODO: How many hashes are no longer needed? Space savings if we delete?
	// TODO: Are there duplicate files?

	return nil
}

// Validate that all three files (orig, meta, and hash) fully match each other.
func verifyMeta(ctx *base.Context, filename string) (bool, error) {
	// Verify we have file metadata.
	meta, err := config.ReadFileMeta(filepath.Join(ctx.GitRepoPath, filename))
	if err != nil {
		e, _ := err.(*os.PathError)
		if e.Err == syscall.ENOENT {
			fmt.Printf("meta doesn't exist... ")
			return false, nil
		}
		return false, err
	}

	// Verify timestamps match.
	info, err := os.Stat(filename)
	if err != nil {
		return false, err
	}

	if meta.UnixNano != info.ModTime().UnixNano() {
		fmt.Printf("file modified... ")
		return false, nil
	}

	// Verify inodes match.
	inode, err := base.GetInode(filename)
	if err != nil {
		return false, err
	}

	inode2, err := base.GetInode(filepath.Join(ctx.HashPath, meta.Sha256))
	if err != nil {
		// If the file doesn't exist that isn't really an error.
		e, _ := err.(*os.PathError)
		if e.Err == syscall.ENOENT {
			fmt.Printf("link missing... ")
			return false, nil
		}
		return false, err
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
		return err
	}

	err = createHardLinkIfNeeded(ctx, filename, hash)
	if err != nil {
		return err
	}

	err = writeFileMeta(ctx, filename, hash)
	if err != nil {
		return err
	}

	return nil
}

func writeFileMeta(ctx *base.Context, filename, sha256 string) error {
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}

	var fm config.FileMeta
	fm.Sha256 = sha256
	fm.UnixNano = info.ModTime().UnixNano()

	d := filepath.Dir(filepath.Join(ctx.GitRepoPath, filename))
	err = os.MkdirAll(d, 0700)
	if err != nil {
		return err
	}

	err = config.WriteFileMeta(ctx.GitRepoPath+"/"+filename, fm)
	if err != nil {
		return err
	}

	return nil
}

func createHardLinkIfNeeded(ctx *base.Context, filename, sha256 string) error {
	// Create a hard link.
	// - If we are already linked is it the correct sha file?
	// - If we aren't already linked, then make it so.
	nlink, err := countHardLinks(filename)
	if err != nil {
		return err
	}

	if nlink == 1 {
		return createHardLink(ctx, filename, sha256)
	}

	inode, err := base.GetInode(filename)
	if err != nil {
		return err
	}

	inode2, err := base.GetInode(ctx.HashPath + "/" + sha256)
	if err != nil {
		return err
	}

	if inode == inode2 {
		fmt.Printf("link exists... ")
		return nil
	}

	existing_hash, err := findInodeHash(ctx, inode)
	if err != nil {
		return err
	}

	if existing_hash != "" {
		if sha256 != existing_hash {
			return fmt.Errorf("corrupted hash file: %s", existing_hash)
		}
		fmt.Printf("link exists... ")
		return nil
	}
	return nil
}

func createHardLink(ctx *base.Context, filename, sha256 string) error {
	err := os.Link(filename, ctx.HashPath+"/"+sha256)
	if err == nil {
		fmt.Printf("link created... ")
		return nil
	}

	e, _ := err.(*os.LinkError)
	if e.Err != syscall.EEXIST {
		return err
	}
	fmt.Print("dup found...")
	fmt.Printf("%s\n", filepath.Join(ctx.DupPath, strings.Replace(filename, "/", "-", -1)))

	os.Rename(filename, filepath.Join(ctx.DupPath, strings.Replace(filename, "/", "-", -1)))

	os.Link(ctx.HashPath+"/"+sha256, filename)

	fmt.Printf("link created... ")
	return nil
}

func findInodeHash(ctx *base.Context, inode uint64) (string, error) {
	var hash string
	err := base.Walk(ctx.HashPath, func(path string, info os.FileInfo) error {
		inode2, e := base.GetInode(path)
		if e != nil {
			return e
		}

		if inode == inode2 {
			hash = filepath.Base(path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return hash, nil
}

func countHardLinks(filename string) (uint64, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return 0, err
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
