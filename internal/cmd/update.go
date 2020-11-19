package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/shortmoose/toobig/internal/base"
	"github.com/shortmoose/toobig/internal/config"
)

var (
	m map[uint64]string
)

// Update the ref and checksum repositories based on the data directory.
func Update(ctx *base.Context) error {
	fmt.Printf("Performing update:\n")

	cfg, err := config.ReadConfig(ctx.ConfigPath)
	if err != nil {
		return err
	}
	ctx.TooBig = cfg

	err = os.Chdir(ctx.DataRepoPath)
	if err != nil {
		return err
	}

	// Create mapping from inode to hashes
	// (files we have already hashed and linked)
	fmt.Printf("Scanning blob directory...\n")
	m = map[uint64]string{}
	err = base.Walk(ctx.HashRepoPath, func(path string, info os.FileInfo) error {
		inode, e := base.GetInode(path)
		if e != nil {
			return e
		}

		m[inode] = filepath.Base(path)
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("Reading data directory...\n")
	// Walk all files in the data directory.
	err = base.Walk(".", func(path string, info os.FileInfo) error {
		er2 := processFile(ctx, path)
		if er2 != nil {
			return er2
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = os.Chdir(ctx.GitRepoPath)
	if err != nil {
		return err
	}

	fmt.Printf("Reading git directory...\n")
	// Walk all files in the data directory.
	err = base.Walk(".", func(path string, info os.FileInfo) error {
		exists, er := base.FileExists(ctx.DataRepoPath + "/" + path)
		if er != nil {
			return er
		}

		if !exists {
			log.Printf("Removing: %s", path)
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
	// TODO: Was there duplicate files?

	return nil
}

func processFile(ctx *base.Context, filename string) error {
	// Try to see if we have already linked this to a file in Hash.
	fmt.Printf("%s... ", filename)
	inode, err := base.GetInode(filename)
	if err != nil {
		return err
	}
	hash, ok := m[inode]

	// Looks like we haven't previously hashed this file.
	if !ok {
		fmt.Printf("hashing... ")
		hash, err = base.GetSha256(filename)
		if err != nil {
			return err
		}

		err = os.Link(filename, ctx.HashRepoPath+"/"+hash)
		if err != nil {
			return err
		}
	}

	// Create a "replica" of the srcRepo in the gitRepo, but each file will just
	// contain the hash of its contents (instead of the actual contents).
	d := filepath.Dir(filepath.Join(ctx.GitRepoPath, filename))
	err = os.MkdirAll(d, 0700)
	if err != nil {
		return err
	}

	var fm config.FileMeta
	fm.Sha256 = hash

	err = config.WriteFileMeta(ctx.GitRepoPath+"/"+filename, fm)
	if err != nil {
		return err
	}

	fmt.Printf("Linked\n")
	return nil
}
