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
	HashRepoPath string `yaml:"blob_path"`
	DataRepoPath string `yaml:"data_path"`
	GitRepoPath  string `yaml:"git_path"`
}

// ReadConfig reads and deserializes TooBig from a file.
func ReadConfig(path string) (TooBig, error) {
	var cfg TooBig

	jsonFile, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return cfg, err
	}

	err = yaml.Unmarshal(byteValue, &cfg)
	if err != nil {
		return cfg, err
	}

	if len(cfg.DataRepoPath) == 0 ||
		len(cfg.HashRepoPath) == 0 ||
		len(cfg.GitRepoPath) == 0 {
		return cfg, fmt.Errorf("Must set data_path, hash_path, and git_path to valid directories")
	}

	p, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return cfg, err
	}

	cfg.GitRepoPath = filepath.Join(p, cfg.GitRepoPath)
	cfg.DataRepoPath = filepath.Join(p, cfg.DataRepoPath)
	cfg.HashRepoPath = filepath.Join(p, cfg.HashRepoPath)

	fmt.Printf("  git_path:  %s/\n", cfg.GitRepoPath)
	fmt.Printf("  data_path: %s/\n", cfg.DataRepoPath)
	fmt.Printf("  hash_path: %s/\n\n", cfg.HashRepoPath)

	return cfg, nil
}
