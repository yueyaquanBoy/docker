// +build windows

package builder

func fixPermissions(source, destination string, uid, gid int, destExisted bool) error {
	// chown not supported on Windows
	return nil
}
