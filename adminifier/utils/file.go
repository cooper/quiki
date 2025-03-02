package utils

import (
	"os"
	"path/filepath"
)

// dirsInDir returns a list of directories in a directory.
// It includes symlinks that point to valid directories.
func DirsInDir(path string) []string {
	files, _ := os.ReadDir(path)
	dirs := make([]string, 0, len(files))
	for _, fi := range files {
		if fi.IsDir() {
			dirs = append(dirs, fi.Name())
		} else if fi.Type()&os.ModeSymlink != 0 {
			// if it's a symlink, follow it and see if it's a directory
			linkPath := filepath.Join(path, fi.Name())
			linkFi, err := os.Stat(linkPath)
			if err == nil && linkFi.IsDir() {
				dirs = append(dirs, fi.Name())
			}
		}
	}
	return dirs
}
