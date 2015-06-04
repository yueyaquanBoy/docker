// +build windows

package daemon

import (
	"fmt"
	"strings"

	"github.com/docker/docker/daemon/execdriver"
)

// TODO Windows. A reasonable default at the moment.
const DefaultPathEnv = `c:\windows\system32;c:\windows\system32\WindowsPowerShell\v1.0`

type Container struct {
	CommonContainer

	// Fields below here are platform specific.

	// TODO Windows. Further factoring out of unused fields will be necessary.

	// ---- START OF TEMPORARY DECLARATION ----
	// TODO Windows. Temporarily keeping fields in to assist in compilation
	// of the daemon on Windows without affecting many other files in a single
	// PR, thus making code review significantly harder. These lines will be
	// removed in subsequent PRs.

	AppArmorProfile string
	// ---- END OF TEMPORARY DECLARATION ----

}

func killProcessDirectly(container *Container) error {
	return nil
}

func (container *Container) setupContainerDns() error {
	return nil
}

func (container *Container) updateParentsHosts() error {
	return nil
}

func (container *Container) setupLinkedContainers() ([]string, error) {
	return nil, nil
}

func (container *Container) createDaemonEnvironment(linkedEnv []string) []string {
	return nil
}

func (container *Container) initializeNetworking() error {
	return nil
}

func (container *Container) setupWorkingDirectory() error {
	return nil
}

func (container *Container) verifyDaemonSettings() {
}

func populateCommand(c *Container, env []string) error {
	en := &execdriver.Network{
		Mtu:       c.daemon.config.Mtu,
		Interface: nil,
	}

	// TODO Windows. Appropriate network mode (will refactor as part of
	// libnetwork. For now, even through bridge not used, let it succeed to
	// allow the Windows daemon to limp during its bring-up
	parts := strings.SplitN(string(c.hostConfig.NetworkMode), ":", 2)
	switch parts[0] {
	case "none":
	case "bridge", "": // empty string to support existing containers
		if !c.Config.NetworkDisabled {
			network := c.NetworkSettings
			en.Interface = &execdriver.NetworkInterface{
				Bridge:     network.Bridge,
				MacAddress: network.MacAddress,
			}
		}
	case "host", "container":
		return fmt.Errorf("unsupported network mode: %s", c.hostConfig.NetworkMode)
	default:
		return fmt.Errorf("invalid network mode: %s", c.hostConfig.NetworkMode)
	}

	pid := &execdriver.Pid{}

	// TODO Windows. This can probably be factored out.
	pid.HostPid = c.hostConfig.PidMode.IsHost()

	// TODO Windows. Resource controls to be implemented later.
	resources := &execdriver.Resources{}

	// TODO Windows. Further refactoring required (privileged/user)
	processConfig := execdriver.ProcessConfig{
		Privileged: c.hostConfig.Privileged,
		Entrypoint: c.Path,
		Arguments:  c.Args,
		Tty:        c.Config.Tty,
		User:       c.Config.User,
	}

	processConfig.Env = env

	// TODO Windows: Factor out remainder of unused fields.
	c.command = &execdriver.Command{
		ID:             c.ID,
		Rootfs:         c.RootfsPath(),
		ReadonlyRootfs: c.hostConfig.ReadonlyRootfs,
		InitPath:       "/.dockerinit",
		WorkingDir:     c.Config.WorkingDir,
		Network:        en,
		Pid:            pid,
		Resources:      resources,
		CapAdd:         c.hostConfig.CapAdd,
		CapDrop:        c.hostConfig.CapDrop,
		ProcessConfig:  processConfig,
		ProcessLabel:   c.GetProcessLabel(),
		MountLabel:     c.GetMountLabel(),
	}

	return nil
}

// GetSize, return real size, virtual size
func (container *Container) GetSize() (int64, int64) {
	// TODO Windows
	return 0, 0
}

func (container *Container) AllocateNetwork() error {

	// TODO Windows. This needs reworking with libnetwork. In the
	// proof-of-concept for //build conference, the Windows daemon
	// invoked eng.Job("allocate_interface) passing through
	// RequestedMac.

	return nil
}

func (container *Container) ReleaseNetwork() {
	// TODO Windows. Rework with libnetwork
}

func (container *Container) RestoreNetwork() error {
	// TODO Windows. Rework with libnetwork
	return nil
}

func disableAllActiveLinks(container *Container) {
}

func (container *Container) DisableLink(name string) {
}

func (container *Container) UnmountVolumes(forceSyscall bool) error {
	return nil
}
