package base

import (
	"os"
	"path/filepath"
)

type WalkFunc func(path string, info os.FileInfo) error

// Walk TODO
func Walk(path string, walkFn WalkFunc) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			base := filepath.Base(path)
			if base != "." && base[0] == '.' {
				return filepath.SkipDir
			}
			return nil
		}

		return walkFn(path, info)
	})
}
