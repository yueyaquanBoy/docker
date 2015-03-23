// +build windows

package execdrivers

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/execdriver"
	"github.com/docker/docker/daemon/execdriver/argon"
	"github.com/docker/docker/daemon/execdriver/windowsdummy"
	"github.com/docker/docker/pkg/sysinfo"
)

func NewDriver(name, root, initPath string, sysInfo *sysinfo.SysInfo) (execdriver.Driver, error) {
	log.Debugln("Windows execdriver - NewDriver ", name)
	switch name {
	case "argon":
		return argon.NewDriver(root, initPath)
	case "windowsdummy":
		return windowsdummy.NewDriver(root, initPath)
	}
	return nil, fmt.Errorf("unknown exec driver %s", name)
}
