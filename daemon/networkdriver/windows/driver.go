// +build windows

package windows

import (
	"net"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/engine"
)

const (
	DefaultNetworkBridge = "Virtual Switch"
)

// Network interface represents the networking stack of a container
type networkInterface struct {
	MACAddress net.HardwareAddr
}

type ifaces struct {
	c map[string]*networkInterface
	sync.Mutex
}

func (i *ifaces) Set(key string, n *networkInterface) {
	i.Lock()
	i.c[key] = n
	i.Unlock()
}

func (i *ifaces) Get(key string) *networkInterface {
	i.Lock()
	res := i.c[key]
	i.Unlock()
	return res
}

var (
	bridgeIface       string
	currentInterfaces = ifaces{c: make(map[string]*networkInterface)}
)

func InitDriver(job *engine.Job) engine.Status {
	log.Debugln("NetworkDriver-Windows Init()")

	if err := SetupMACRange([]byte{0x02, 0x42}); err != nil {
		return job.Error(err)
	}

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

	bridgeIface = job.Getenv("BridgeIface")
	if bridgeIface == "" {
		bridgeIface = DefaultNetworkBridge
	}

	return engine.StatusOK
}

// Allocate a network interface
func Allocate(job *engine.Job) engine.Status {
	log.Debugln("NetworkDriver-Windows Allocate()")

	var (
		mac net.HardwareAddr
		id  = job.Args[0]
		err error
	)

	out := engine.Env{}
	out.Set("Bridge", bridgeIface)

	// If no explicit mac address was given, generate a random one.
	if mac, err = net.ParseMAC(job.Getenv("RequestedMac")); err != nil {
		if mac, err = RequestMAC(); err != nil {
			return job.Error(err)
		}
	}

	out.Set("MacAddress", mac.String())
	log.Debugln("NetworkDriver-Windows MAC=", mac.String())

	// If no explicit mac address was given, generate a random one.
	if mac, err = net.ParseMAC(job.Getenv("RequestedMac")); err != nil {
		if mac, err = RequestMAC(); err != nil {
			return job.Error(err)
		}
	}

	out.Set("MacAddress", mac.String())
	log.Debugln("NetworkDriver-Windows MAC=", mac.String())

	out.WriteTo(job.Stdout)

	currentInterfaces.Set(id, &networkInterface{
		MACAddress: mac,
	})

	return engine.StatusOK
}

// release an interface for a select ip
func Release(job *engine.Job) engine.Status {
	log.Debugln("NetworkDriver-Windows Release()")
	var (
		id                 = job.Args[0]
		containerInterface = currentInterfaces.Get(id)
	)
	if containerInterface == nil {
		return job.Errorf("No network information to release for %s", id)
	}
	if err := ReleaseMac(containerInterface.MACAddress); err != nil {
		log.Infof("Unable to release MAC Address %s %s", containerInterface.MACAddress.String(), err)
	}
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
