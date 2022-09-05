package util

import (
	"os"
	"path/filepath"
)

// ReadFile a safe wrapper of os.ReadFile.
func ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filepath.Clean(filename))
}
