package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// TooBig represents a TooBig repository configuration.
type TooBig struct {
	FilePath string `json:"file-path"`
	BlobPath string `json:"blob-path"`
	RefPath  string `json:"ref-path"`
	DupPath  string `json:"dup-path"`
}

func ReadConfig(path string) (TooBig, error) {
	cfg, err := readConfig(path)
	if err != nil {
		return cfg, fmt.Errorf("reading config file %s: %w", path, err)
	}

	if !is_dir(cfg.FilePath) {
		return cfg, fmt.Errorf("file-path, %s, is not a directory", cfg.FilePath)
	}
	if !is_dir(cfg.BlobPath) {
		return cfg, fmt.Errorf("blob-path, %s, is not a directory", cfg.BlobPath)
	}
	if !is_dir(cfg.RefPath) {
		return cfg, fmt.Errorf("ref-path, %s, is not a directory", cfg.RefPath)
	}
	if !is_dir(cfg.DupPath) {
		return cfg, fmt.Errorf("dup-path, %s, is not a directory", cfg.DupPath)
	}

	return cfg, nil
}

// ReadConfig reads and deserializes TooBig from a file.
func readConfig(path string) (TooBig, error) {
	var cfg TooBig

	jsonFile, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer func() { _ = jsonFile.Close() }() // Yes, we are ignoring errors

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return cfg, err
	}

	err = json.Unmarshal(byteValue, &cfg)
	if err != nil {
		return cfg, err
	}

	if cfg.FilePath == "" ||
		cfg.BlobPath == "" ||
		cfg.RefPath == "" ||
		cfg.DupPath == "" {
		return cfg, fmt.Errorf("file-path, blob-path, ref-path, and dup-path must be valid directories")
	}

	p, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return cfg, err
	}

	join := func(root, add string) string {
		if filepath.IsAbs(add) {
			return filepath.Clean(add)
		}
		return filepath.Clean(filepath.Join(root, add))
	}

	cfg.RefPath = join(p, cfg.RefPath)
	cfg.FilePath = join(p, cfg.FilePath)
	cfg.BlobPath = join(p, cfg.BlobPath)
	cfg.DupPath = join(p, cfg.DupPath)

	fmt.Printf("  file-path: %s\n", cfg.FilePath)
	fmt.Printf("  ref-path:  %s\n", cfg.RefPath)
	fmt.Printf("  blob-path: %s\n", cfg.BlobPath)
	fmt.Printf("  dup-path:  %s\n\n", cfg.DupPath)

	return cfg, nil
}

func is_dir(dir string) bool {
	fileInfo, err := os.Stat(dir)
	if err != nil || !fileInfo.IsDir() {
		return false
	}

	return true
}

func WriteExampleConfig() {
	var cfg TooBig
	cfg.FilePath = "files"
	cfg.BlobPath = "blobs"
	cfg.RefPath = "refs"
	cfg.DupPath = "dups"

	out, err := json.MarshalIndent(cfg, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}
