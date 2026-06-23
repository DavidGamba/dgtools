package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func createCacheDir(subDir string) (fullpath string, err error) {
	cacheDirBase, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user cache dir: %w", err)
	}
	cacheDir := filepath.Join(cacheDirBase, "jsql", subDir)
	os.MkdirAll(cacheDir, 0755)
	return cacheDir, nil
}
