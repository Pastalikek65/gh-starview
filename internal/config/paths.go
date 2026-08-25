package config

import (
	"os"
	"path/filepath"
)

func CacheDir() string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "gh-starview")
	}
	home, _ := os.UserHomeDir()
	if home == "" {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".cache", "gh-starview")
}

func CacheDBPath() string {
	return filepath.Join(CacheDir(), "gh-starview.db")
}

func EnsureCacheDir() error {
	return os.MkdirAll(CacheDir(), 0755)
}
