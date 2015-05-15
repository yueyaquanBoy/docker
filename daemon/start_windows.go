// +build windows

package daemon

import (
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/graphdriver/windows"
	"github.com/docker/docker/pkg/hcsshim"
)

func (daemon *Daemon) setupStorage(container *Container) error {
	if wd, ok := daemon.driver.(*windows.WindowsGraphDriver); ok {
		// Get list of paths to parent layers.
		log.Debugln("Container has parent image:", container.ImageID)
		img, err := daemon.graph.Get(container.ImageID)
		if err != nil {
			return err
		}

		ids, err := daemon.graph.ParentLayerIds(img)
		if err != nil {
			return err
		}
		log.Debugln("Got image ids:", len(ids))

		return hcsshim.PrepareLayer(wd.Info(), container.ID, wd.LayerIdsToPaths(ids))
	}

	return nil
}
