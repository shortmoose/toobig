package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// FileMeta is used to serialize metadata for a file.
type FileMeta struct {
	Filename string `json:"filename"`
	Sha256   string `json:"sha256"`
	Unixtime string `json:"unixtime"`
}

// ReadFileMeta reads and deserializes FileMeta from a file.
func ReadFileMeta(path string) (FileMeta, error) {
	var fm FileMeta

	f, err := os.Open(path)
	if err != nil {
		return fm, err
	}
	defer f.Close()

	sha256, err := ioutil.ReadAll(f)
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

	return ioutil.WriteFile(path, file, 0644)
}
