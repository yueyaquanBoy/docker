// +build daemon

package utils

// IsFileOwner checks whether the current user is the owner of the given file.
func IsFileOwner(f string) bool {
	// TODO Windows. Implement this, or deprecate. TBD.
	return false
}
