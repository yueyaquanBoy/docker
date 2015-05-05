// +build linux

package daemon

import (
	"strings"

	"github.com/docker/docker/daemon/execdriver/lxc"
)

func lxcCheck(DriverName string) error {
	if strings.HasPrefix(DriverName, lxc.DriverName) {
		return lxc.ErrExec
	}
	return nil
}
