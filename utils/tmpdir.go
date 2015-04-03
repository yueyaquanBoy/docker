package utils

import (
	"os"
	"path/filepath"

	"github.com/docker/docker/pkg/system"
)

// TempDir returns the default directory to use for temporary files.
func TempDir(rootDir string) (string, error) {
	var tmpDir string
	if tmpDir = os.Getenv("DOCKER_TMPDIR"); tmpDir == "" {
		tmpDir = filepath.Join(rootDir, "tmp")
	}
	err := system.MkdirAll(tmpDir, 0700)
	return tmpDir, err
}
