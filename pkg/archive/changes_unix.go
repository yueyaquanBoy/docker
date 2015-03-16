// +build !windows

package archive

import (
	"syscall"

	"github.com/docker/docker/pkg/system"
)

type FileInfo struct {
	parent     *FileInfo
	name       string
	stat       *system.Stat_t
	children   map[string]*FileInfo
	capability []byte
	added      bool
}

func (info *FileInfo) isDir() bool {
	return info.parent == nil || info.stat.Mode()&syscall.S_IFDIR == syscall.S_IFDIR
}

type Stat struct {
	mode uint32
	uid  uint32
	gid  uint32
	rdev uint64
	size int64
	mtim syscall.Timespec
}

func statDifferent(oldStat *system.Stat_t, newStat *system.Stat_t) bool {
	// Don't look at size for dirs, its not a good measure of change
	if oldStat.Mode() != newStat.Mode() ||
		oldStat.Uid() != newStat.Uid() ||
		oldStat.Gid() != newStat.Gid() ||
		oldStat.Rdev() != newStat.Rdev() ||
		(oldStat.Size() != newStat.Size() && oldStat.Mode()&syscall.S_IFDIR != syscall.S_IFDIR) ||
		!sameFsTimeSpec(oldStat.Mtim(), newStat.Mtim()) {
		return true
	} else {
		return false
	}
}
