// +build windows

package system

import (
	"os"
	"time"
)

type Stat struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (s Stat) Name() string {
	return s.name
}

func (s Stat) Size() int64 {
	return s.size
}

func (s Stat) Mode() os.FileMode {
	return s.mode
}

func (s Stat) ModTime() time.Time {
	return s.modTime
}

func (s Stat) IsDir() bool {
	return s.isDir
}

func Stat(path string) (*Stat_t, error) {
	// should not be called on cli code path
	return nil, ErrNotSupportedPlatform
}
