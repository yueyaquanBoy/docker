package archive

import (
	"syscall"

	"github.com/docker/docker/pkg/system"
)

func (info *FileInfo) isDir() bool {
	return info.parent == nil || info.stat.IsDir()
	// TODO WINDOWS JJH Think I want the above line. Verify.
	//return info.parent == nil || info.stat.FileAttributes()&syscall.FILE_ATTRIBUTE_DIRECTORY == syscall.FILE_ATTRIBUTE_DIRECTORY
}

func statDifferent(oldStat *system.Stat, newStat *system.Stat) bool {
	// Don't look at size for dirs, its not a good measure of change
	if oldStat.CreationTime() != newStat.CreationTime() ||
		oldStat.FileAttributes() != newStat.FileAttributes() ||
		oldStat.FileSizeHigh() != newStat.FileSizeHigh() && oldStat.FileAttributes()&syscall.FILE_ATTRIBUTE_DIRECTORY != syscall.FILE_ATTRIBUTE_DIRECTORY ||
		oldStat.FileSizeLow() != newStat.FileSizeLow() && oldStat.FileAttributes()&syscall.FILE_ATTRIBUTE_DIRECTORY != syscall.FILE_ATTRIBUTE_DIRECTORY ||
		oldStat.LastWriteTime().HighDateTime != oldStat.LastWriteTime().HighDateTime ||
		oldStat.LastWriteTime().LowDateTime != oldStat.LastWriteTime().LowDateTime {
		return true
	} else {
		return false
	}
}
