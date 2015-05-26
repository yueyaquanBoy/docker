// +build include_graphdriver_vfs

package graphdriver

// This file enables the use of the VFS driver by build-tag for development of
// the docker daemon on Windows without using Windows containers. VFS is not
// capable of backing actual Windows Server containers or Hyper-V containers.

var (
	// Slice of drivers that should be used in an order
	priority = []string{
		"windows",
		"vfs",
	}
)
