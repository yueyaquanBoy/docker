// +build windows

package system

import (
	"os"
)

// Some explanation from my own sanity while getting this going.
// Trust me, this has taken hours to work out the crazy factoring.
//
// Lstat calls os.Lstat to get a fileinfo interface back.
// This is then copied into our own locally defined structure.
// Note the Linux version uses fromStatT to do the copy back,
// but that seems like a lot of overkill to me.
//
// JJH March 2015 as part of the Windows docker daemon port
func Lstat(path string) (*Stat, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	// Need to populate one of these:
	//type Stat struct {
	//	name    string
	//	size    int64
	//	mode    os.FileMode
	//	modTime time.Time
	//	isDir   bool
	//}

	return &Stat{name: fi.Name(),
		size:    fi.Size(),
		mode:    fi.Mode(),
		modTime: fi.ModTime(),
		isDir:   fi.IsDir()}, nil
}
