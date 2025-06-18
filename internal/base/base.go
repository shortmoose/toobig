package base

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"syscall"
)

// GetInode returns the inode of the given file.
func GetInode(filename string) (uint64, error) {
	fileinfo, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}

	stat, ok := fileinfo.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, fmt.Errorf("unable to get syscall.stat_t")
	}
	return stat.Ino, nil
}

// GetSha256 computes and returns the SHA256 for the given file.
func GetSha256(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }() // Explicitly ignore the error

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// FileExists determines if the given file exists (permissions allowing).
func FileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil

	} else if os.IsNotExist(err) {
		return false, nil

	} else {
		return false, err
	}
}
