// +build windows

package archive

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/system"
	"github.com/docker/docker/vendor/src/code.google.com/p/go/src/pkg/archive/tar"
)

// canonicalTarNameForPath returns platform-specific filepath
// to canonical posix-style path for tar archival. p is relative
// path.
func CanonicalTarNameForPath(p string) (string, error) {
	// windows: convert windows style relative path with backslashes
	// into forward slashes. since windows does not allow '/' or '\'
	// in file names, it is mostly safe to replace however we must
	// check just in case
	if strings.Contains(p, "/") {
		return "", fmt.Errorf("windows path contains forward slash: %s", p)
	}
	return strings.Replace(p, string(os.PathSeparator), "/", -1), nil

}

// chmodTarEntry is used to adjust the file permissions used in tar header based
// on the platform the archival is done.
func chmodTarEntry(perm os.FileMode) os.FileMode {
	// Clear r/w on grp/others: no precise equivalen of group/others on NTFS.
	perm &= 0711
	// Add the x bit: make everything +x from windows
	perm |= 0100

	return perm
}

func setHeaderForSpecialDevice(hdr *tar.Header, ta *tarAppender, name string, stat interface{}) (nlink uint32, inode uint64, err error) {
	// do nothing. no notion of Rdev, Inode, Nlink in stat on Windows
	return
}

// Note Lchown parameter is ignored on Windows as not supported
func createTarFile(path, extractDir string, hdr *tar.Header, reader io.Reader, Lchown bool) error {
	// hdr.Mode is in linux format, which we can use for sycalls,
	// but for os.Foo() calls we need the mode converted to os.FileMode,
	// so use hdrInfo.Mode() (they differ for e.g. setuid bits)
	hdrInfo := hdr.FileInfo()

	switch hdr.Typeflag {
	case tar.TypeDir:
		// Create directory unless it exists as a directory already.
		// In that case we just want to merge the two
		if fi, err := os.Lstat(path); !(err == nil && fi.IsDir()) {
			if err := os.Mkdir(path, hdrInfo.Mode()); err != nil {
				return err
			}
		}

	case tar.TypeReg, tar.TypeRegA:
		// Source is regular file
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, hdrInfo.Mode())
		if err != nil {
			return err
		}
		if _, err := io.Copy(file, reader); err != nil {
			file.Close()
			return err
		}
		if err := file.Sync(); err != nil {
			return err
		}
		file.Close()

	case tar.TypeBlock, tar.TypeChar, tar.TypeFifo:
		// Mknod is not supported on Windows

	case tar.TypeLink:
		targetPath := filepath.Join(extractDir, hdr.Linkname)
		// check for hardlink breakout
		if !strings.HasPrefix(targetPath, extractDir) {
			return breakoutError(fmt.Errorf("invalid hardlink %q -> %q", targetPath, hdr.Linkname))
		}
		if err := os.Link(targetPath, path); err != nil {
			return err
		}

	case tar.TypeSymlink:
		// Symbolic link not supported on Windows. However do validation.

		// 	path 				-> hdr.Linkname = targetPath
		// e.g. /extractDir/path/to/symlink 	-> ../2/file	= /extractDir/path/2/file
		targetPath := filepath.Join(filepath.Dir(path), hdr.Linkname)

		// the reason we don't need to check symlinks in the path (with FollowSymlinkInScope) is because
		// that symlink would first have to be created, which would be caught earlier, at this very check:
		if !strings.HasPrefix(targetPath, extractDir) {
			return breakoutError(fmt.Errorf("invalid symlink %q -> %q", path, hdr.Linkname))
		}

	case tar.TypeXGlobalHeader:
		log.Debugf("PAX Global Extended Headers found and ignored")
		return nil

	default:
		return fmt.Errorf("Unhandled tar header type %d\n", hdr.Typeflag)
	}

	for key, value := range hdr.Xattrs {
		if err := system.Lsetxattr(path, key, []byte(value), 0); err != nil {
			return err
		}
	}

	ts := []syscall.Timespec{timeToTimespec(hdr.AccessTime), timeToTimespec(hdr.ModTime)}
	// syscall.UtimesNano doesn't support a NOFOLLOW flag atm, and
	if hdr.Typeflag != tar.TypeSymlink {
		if err := system.UtimesNano(path, ts); err != nil && err != system.ErrNotSupportedPlatform {
			return err
		}
	} else {
		if err := system.LUtimesNano(path, ts); err != nil && err != system.ErrNotSupportedPlatform {
			return err
		}
	}
	return nil
}
