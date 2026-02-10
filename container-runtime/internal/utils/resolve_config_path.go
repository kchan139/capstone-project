package utils

import (
	"os"
	"path/filepath"
)

func ResolveConfigPath(bundlePath string) (string, error) {
	if bundlePath != "" {
		return filepath.Join(bundlePath, "config.json"), nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Join(cwd, "config.json"), nil
}


func ResolvePath(p, bundlePath string) string {
    if filepath.IsAbs(p) {
        return filepath.Clean(p)
    }

    return filepath.Clean(filepath.Join(bundlePath, p))
}