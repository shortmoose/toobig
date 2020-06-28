package base

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
)

func GetInode(filename string) (uint64, error) {
	fileinfo, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}

	stat, ok := fileinfo.Sys().(*syscall.Stat_t)
	if !ok {
		log.Fatal("Hmmm")
	}
	return stat.Ino, nil
}

// GetSha256 TODO
func GetSha256(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func FileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil

	} else if os.IsNotExist(err) {
		return false, nil

	} else {
		return false, err
	}
}
