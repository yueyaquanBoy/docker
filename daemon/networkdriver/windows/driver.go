package windows

import (
	log "github.com/Sirupsen/logrus"
	//"github.com/docker/docker/daemon/networkdriver"
	"github.com/docker/docker/engine"
)

func InitDriver(job *engine.Job) engine.Status {
	log.Debugln("NetworkDriver-Windows Init()")
	for name, f := range map[string]engine.Handler{
		"allocate_interface": Allocate,
		"release_interface":  Release,
		"allocate_port":      AllocatePort,
		"link":               LinkContainers,
	} {
		if err := job.Eng.Register(name, f); err != nil {
			return job.Error(err)
		}
	}
	return engine.StatusOK
}

// Allocate a network interface
func Allocate(job *engine.Job) engine.Status {
	log.Debugln("NetworkDriver-Windows Allocate()")
	return engine.StatusOK
}

// release an interface for a select ip
func Release(job *engine.Job) engine.Status {
	log.Debugln("NetworkDriver-Windows Release()")
	return engine.StatusOK
}

// Allocate an external port and map it to the interface
func AllocatePort(job *engine.Job) engine.Status {
	log.Debugln("NetworkDriver-Windows AllocatePort()")
	return engine.StatusOK
}

func LinkContainers(job *engine.Job) engine.Status {
	log.Debugln("NetworkDriver-Windows LinkContainers()")
	return engine.StatusOK
}
