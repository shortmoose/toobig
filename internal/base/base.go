package base

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

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
