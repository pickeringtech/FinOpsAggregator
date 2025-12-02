package storage

import (
	"os"
)

// ensureDir creates a directory if it doesn't exist.
func ensureDir(path string) error {
	if path == "" {
		return nil
	}
	return os.MkdirAll(path, 0755)
}

