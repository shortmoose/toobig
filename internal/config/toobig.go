package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// TooBig represents a TooBig repository configuration.
type TooBig struct {
	HashPath    string `yaml:"blob_path"`
	DataPath    string `yaml:"data_path"`
	GitRepoPath string `yaml:"git_path"`
	DupPath     string `yaml:"dup_path"`
}

// ReadConfig reads and deserializes TooBig from a file.
func ReadConfig(path string) (TooBig, error) {
	var cfg TooBig

	jsonFile, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer func() { _ = jsonFile.Close() }() // Explicitly ignore the error

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return cfg, err
	}

	err = yaml.Unmarshal(byteValue, &cfg)
	if err != nil {
		return cfg, err
	}

	if cfg.DataPath == "" ||
		cfg.HashPath == "" ||
		cfg.GitRepoPath == "" ||
		cfg.DupPath == "" {
		return cfg, fmt.Errorf("Must set data_path, hash_path, git_path, and dup_path to valid directories")
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

	cfg.GitRepoPath = join(p, cfg.GitRepoPath)
	cfg.DataPath = join(p, cfg.DataPath)
	cfg.HashPath = join(p, cfg.HashPath)
	cfg.DupPath = join(p, cfg.DupPath)

	fmt.Printf("  git_path:  %s\n", cfg.GitRepoPath)
	fmt.Printf("  data_path: %s\n", cfg.DataPath)
	fmt.Printf("  hash_path: %s\n", cfg.HashPath)
	fmt.Printf("  dup_path:  %s\n\n", cfg.DupPath)

	return cfg, nil
}
