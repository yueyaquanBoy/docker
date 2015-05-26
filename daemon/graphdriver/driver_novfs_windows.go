// +build !include_graphdriver_vfs

package graphdriver

var (
	// Slice of drivers that should be used in an order
	priority = []string{
		"windows",
	}
)
