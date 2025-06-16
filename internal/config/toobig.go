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
	FilePath string `json:"file_path"`
	BlobPath string `json:"blob_path"`
	RefPath  string `json:"ref_path"`
	DupPath  string `json:"dup_path"`
}

func ReadConfig(path string) (TooBig, error) {
	cfg, err := readConfig(path)
	if err != nil {
		return cfg, fmt.Errorf("reading config file %s: %w", path, err)
	}
	return cfg, err
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
		return cfg, fmt.Errorf("file_path, blob_path, ref_path, and dup_path must be valid directories")
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

	fmt.Printf("  file_path: %s\n", cfg.FilePath)
	fmt.Printf("  ref_path:  %s\n", cfg.RefPath)
	fmt.Printf("  blob_path: %s\n", cfg.BlobPath)
	fmt.Printf("  dup_path:  %s\n\n", cfg.DupPath)

	return cfg, nil
}

func WriteExampleConfig() {
	var cfg TooBig

	out, err := json.MarshalIndent(cfg, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}
