// +build windows

package system

import (
	"syscall"
)

type Stat struct {
	fileAttributes uint32
	creationTime   syscall.Filetime
	lastWriteTime  syscall.Filetime
	fileSizeHigh   uint32
	fileSizeLow    uint32
}

func (s Stat) FileAttributes() uint32 {
	return s.fileAttributes
}

func (s Stat) CreationTime() syscall.Filetime {
	return s.creationTime
}

func (s Stat) LastWriteTime() syscall.Filetime {
	return s.lastWriteTime
}

func (s Stat) FileSizeHigh() uint32 {
	return s.fileSizeHigh
}

func (s Stat) FileSizeLow() uint32 {
	return s.fileSizeLow
}

func fromStatT(s *syscall.Win32FileAttributeData) (*Stat, error) {
	return &Stat{fileAttributes: s.FileAttributes,
		creationTime:  s.CreationTime,
		lastWriteTime: s.LastWriteTime,
		fileSizeHigh:  s.FileSizeHigh,
		fileSizeLow:   s.FileSizeLow}, nil
}

func Stat(path string) (*Stat_t, error) {
	// should not be called on cli code path
	return nil, ErrNotSupportedPlatform
}
