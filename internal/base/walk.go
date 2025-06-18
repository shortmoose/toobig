package base

import (
	"io/fs"
	"os"
	"path/filepath"
)

type WalkFunc func(path string, info fs.DirEntry) error

// Walk the given path, mostly a simple wrapper around filepath.Walk.
func Walk(path string, walkFn WalkFunc) error {
	return filepath.WalkDir(path, func(path string, info fs.DirEntry, err error) error {
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

// Walk the given path, but chdir first.
func ChdirWalk(path string, walkFn WalkFunc) error {
	err := os.Chdir(path)
	if err != nil {
		return err
	}

	return Walk(".", walkFn)
}
