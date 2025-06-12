package config

import (
	"encoding/json"
	"io"
	"os"
)

// FileMeta is used to serialize metadata for a file.
type FileMeta struct {
	Sha256   string `json:"sha256"`
	UnixNano int64  `json:"unixnano"`
}

// ReadFileMeta reads and deserializes FileMeta from a file.
func ReadFileMeta(path string) (FileMeta, error) {
	var fm FileMeta

	f, err := os.Open(path)
	if err != nil {
		return fm, err
	}
	defer func() { _ = f.Close() }() // Explicitly ignore the error

	sha256, err := io.ReadAll(f)
	if err != nil {
		return fm, err
	}

	err = json.Unmarshal(sha256, &fm)
	if err != nil {
		return fm, err
	}

	return fm, nil
}

// WriteFileMeta serializes and writes FileMeta to a file.
func WriteFileMeta(path string, fm FileMeta) error {
	file, err := json.MarshalIndent(fm, "", " ")
	if err != nil {
		return err
	}

	file = append(file, '\n')

	return os.WriteFile(path, file, 0644)
}
