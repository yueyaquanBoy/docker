package graph

import (
	"os"
	"path"
	"strings"
	"syscall"
)

// setupInitLayer populates a directory with mountpoints suitable
// for bind-mounting dockerinit into the container. The mountpoint is simply an
// empty file at /.dockerinit
//
// This extra layer is used by all containers as the top-most ro layer. It protects
// the container from unwanted side-effects on the rw layer.
func SetupInitLayer(initLayer string) error {
	for pth, typ := range map[string]string{
		"/dev/pts":         "dir",
		"/dev/shm":         "dir",
		"/proc":            "dir",
		"/sys":             "dir",
		"/.dockerinit":     "file",
		"/.dockerenv":      "file",
		"/etc/resolv.conf": "file",
		"/etc/hosts":       "file",
		"/etc/hostname":    "file",
		"/dev/console":     "file",
		"/etc/mtab":        "/proc/mounts",
	} {
		parts := strings.Split(pth, "/")
		prev := "/"
		for _, p := range parts[1:] {
			prev = path.Join(prev, p)
			syscall.Unlink(path.Join(initLayer, prev))
		}

		if _, err := os.Stat(path.Join(initLayer, pth)); err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(path.Join(initLayer, path.Dir(pth)), 0755); err != nil {
					return err
				}
				switch typ {
				case "dir":
					if err := os.MkdirAll(path.Join(initLayer, pth), 0755); err != nil {
						return err
					}
				case "file":
					f, err := os.OpenFile(path.Join(initLayer, pth), os.O_CREATE, 0755)
					if err != nil {
						return err
					}
					f.Close()
				default:
					if err := os.Symlink(typ, path.Join(initLayer, pth)); err != nil {
						return err
					}
				}
			} else {
				return err
			}
		}
	}

	// Layer is ready to use, if it wasn't before.
	return nil
}
