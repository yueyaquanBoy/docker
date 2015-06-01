// +build !windows

package daemon

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/autogen/dockerversion"
	"github.com/docker/docker/daemon/graphdriver"
	_ "github.com/docker/docker/daemon/graphdriver/vfs"
	"github.com/docker/docker/graph"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/docker/docker/pkg/parsers/kernel"
	"github.com/docker/docker/runconfig"
	"github.com/docker/docker/utils"
	volumedrivers "github.com/docker/docker/volume/drivers"
	"github.com/docker/docker/volume/local"
	"github.com/docker/libcontainer/label"
	"github.com/docker/libnetwork"
	"github.com/docker/libnetwork/netlabel"
	"github.com/docker/libnetwork/options"
)

const runDir = "/var/run/docker"
const defaultVolumesPathName = "volumes"

func (daemon *Daemon) Changes(container *Container) ([]archive.Change, error) {
	initID := fmt.Sprintf("%s-init", container.ID)
	return daemon.driver.Changes(container.ID, initID)
}

func (daemon *Daemon) Diff(container *Container) (archive.Archive, error) {
	initID := fmt.Sprintf("%s-init", container.ID)
	return daemon.driver.Diff(container.ID, initID)
}

func parseSecurityOpt(container *Container, config *runconfig.HostConfig) error {
	var (
		labelOpts []string
		err       error
	)

	for _, opt := range config.SecurityOpt {
		con := strings.SplitN(opt, ":", 2)
		if len(con) == 1 {
			return fmt.Errorf("Invalid --security-opt: %q", opt)
		}
		switch con[0] {
		case "label":
			labelOpts = append(labelOpts, con[1])
		case "apparmor":
			container.AppArmorProfile = con[1]
		default:
			return fmt.Errorf("Invalid --security-opt: %q", opt)
		}
	}

	container.ProcessLabel, container.MountLabel, err = label.InitLabels(labelOpts)
	return err
}

func (daemon *Daemon) createRootfs(container *Container) error {
	// Step 1: create the container directory.
	// This doubles as a barrier to avoid race conditions.
	if err := os.Mkdir(container.root, 0700); err != nil {
		return err
	}
	initID := fmt.Sprintf("%s-init", container.ID)
	if err := daemon.driver.Create(initID, container.ImageID); err != nil {
		return err
	}
	initPath, err := daemon.driver.Get(initID, "")
	if err != nil {
		return err
	}
	defer daemon.driver.Put(initID)

	if err := graph.SetupInitLayer(initPath); err != nil {
		return err
	}

	if err := daemon.driver.Create(container.ID, initID); err != nil {
		return err
	}
	return nil
}

func checkKernel() error {
	// Check for unsupported kernel versions
	// FIXME: it would be cleaner to not test for specific versions, but rather
	// test for specific functionalities.
	// Unfortunately we can't test for the feature "does not cause a kernel panic"
	// without actually causing a kernel panic, so we need this workaround until
	// the circumstances of pre-3.10 crashes are clearer.
	// For details see https://github.com/docker/docker/issues/407
	if k, err := kernel.GetKernelVersion(); err != nil {
		logrus.Warnf("%s", err)
	} else {
		if kernel.CompareKernelVersion(k, &kernel.KernelVersionInfo{Kernel: 3, Major: 10, Minor: 0}) < 0 {
			if os.Getenv("DOCKER_NOWARN_KERNEL_VERSION") == "" {
				logrus.Warnf("You are running linux kernel version %s, which might be unstable running docker. Please upgrade your kernel to 3.10.0.", k.String())
			}
		}
	}
	return nil
}

func (daemon *Daemon) verifyHostConfig(hostConfig *runconfig.HostConfig) ([]string, error) {
	var warnings []string

	if hostConfig == nil {
		return warnings, nil
	}

	if hostConfig.LxcConf.Len() > 0 && !strings.Contains(daemon.ExecutionDriver().Name(), "lxc") {
		return warnings, fmt.Errorf("Cannot use --lxc-conf with execdriver: %s", daemon.ExecutionDriver().Name())
	}
	if hostConfig.Memory != 0 && hostConfig.Memory < 4194304 {
		return warnings, fmt.Errorf("Minimum memory limit allowed is 4MB")
	}
	if hostConfig.Memory > 0 && !daemon.SystemConfig().MemoryLimit {
		warnings = append(warnings, "Your kernel does not support memory limit capabilities. Limitation discarded.")
		hostConfig.Memory = 0
	}
	if hostConfig.Memory > 0 && hostConfig.MemorySwap != -1 && !daemon.SystemConfig().SwapLimit {
		warnings = append(warnings, "Your kernel does not support swap limit capabilities, memory limited without swap.")
		hostConfig.MemorySwap = -1
	}
	if hostConfig.Memory > 0 && hostConfig.MemorySwap > 0 && hostConfig.MemorySwap < hostConfig.Memory {
		return warnings, fmt.Errorf("Minimum memoryswap limit should be larger than memory limit, see usage.")
	}
	if hostConfig.Memory == 0 && hostConfig.MemorySwap > 0 {
		return warnings, fmt.Errorf("You should always set the Memory limit when using Memoryswap limit, see usage.")
	}
	if hostConfig.CpuPeriod > 0 && !daemon.SystemConfig().CpuCfsPeriod {
		warnings = append(warnings, "Your kernel does not support CPU cfs period. Period discarded.")
		hostConfig.CpuPeriod = 0
	}
	if hostConfig.CpuQuota > 0 && !daemon.SystemConfig().CpuCfsQuota {
		warnings = append(warnings, "Your kernel does not support CPU cfs quota. Quota discarded.")
		hostConfig.CpuQuota = 0
	}
	if hostConfig.BlkioWeight > 0 && (hostConfig.BlkioWeight < 10 || hostConfig.BlkioWeight > 1000) {
		return warnings, fmt.Errorf("Range of blkio weight is from 10 to 1000.")
	}
	if hostConfig.OomKillDisable && !daemon.SystemConfig().OomKillDisable {
		hostConfig.OomKillDisable = false
		return warnings, fmt.Errorf("Your kernel does not support oom kill disable.")
	}

	return warnings, nil
}

// checkConfigOptions checks for mutually incompatible config options
func checkConfigOptions(config *Config) error {
	// Check for mutually incompatible config options
	if config.Bridge.Iface != "" && config.Bridge.IP != "" {
		return fmt.Errorf("You specified -b & --bip, mutually exclusive options. Please specify only one.")
	}
	if !config.Bridge.EnableIPTables && !config.Bridge.InterContainerCommunication {
		return fmt.Errorf("You specified --iptables=false with --icc=false. ICC uses iptables to function. Please set --icc or --iptables to true.")
	}
	if !config.Bridge.EnableIPTables && config.Bridge.EnableIPMasq {
		config.Bridge.EnableIPMasq = false
	}
	return nil
}

// checkSystem validates the system is supported and we have sufficient privileges
func checkSystem() error {
	// TODO Windows. Once daemon is running on Windows, move this code back to
	// NewDaemon() in daemon.go, and extend the check to support Windows.
	if runtime.GOOS != "linux" {
		return fmt.Errorf(ErrSystemNotSupported)
	}
	if os.Geteuid() != 0 {
		return fmt.Errorf("The Docker daemon needs to be run as root")
	}
	if err := checkKernel(); err != nil {
		return err
	}
	return nil
}

// configureKernelSecuritySupport configures and validate security support for the kernel
func configureKernelSecuritySupport(config *Config, driverName string) error {
	if config.EnableSelinuxSupport {
		if selinuxEnabled() {
			// As Docker on btrfs and SELinux are incompatible at present, error on both being enabled
			if driverName == "btrfs" {
				return fmt.Errorf("SELinux is not supported with the BTRFS graph driver")
			}
			logrus.Debug("SELinux enabled successfully")
		} else {
			logrus.Warn("Docker could not enable SELinux on the host system")
		}
	} else {
		selinuxSetDisabled()
	}
	return nil
}

// MigrateIfDownlevel is a wrapper for AUFS migration for downlevel
func migrateIfDownlevel(driver graphdriver.Driver, root string) error {
	return migrateIfAufs(driver, root)
}

func configureVolumes(config *Config) error {
	volumesDriver, err := local.New(filepath.Join(config.Root, defaultVolumesPathName))
	if err != nil {
		return err
	}
	volumedrivers.Register(volumesDriver, volumesDriver.Name())
	return nil
}

func configureSysInit(config *Config) (string, error) {
	localCopy := filepath.Join(config.Root, "init", fmt.Sprintf("dockerinit-%s", dockerversion.VERSION))
	sysInitPath := utils.DockerInitPath(localCopy)
	if sysInitPath == "" {
		return "", fmt.Errorf("Could not locate dockerinit: This usually means docker was built incorrectly. See https://docs.docker.com/contributing/devenvironment for official build instructions.")
	}

	if sysInitPath != localCopy {
		// When we find a suitable dockerinit binary (even if it's our local binary), we copy it into config.Root at localCopy for future use (so that the original can go away without that being a problem, for example during a package upgrade).
		if err := os.Mkdir(filepath.Dir(localCopy), 0700); err != nil && !os.IsExist(err) {
			return "", err
		}
		if _, err := fileutils.CopyFile(sysInitPath, localCopy); err != nil {
			return "", err
		}
		if err := os.Chmod(localCopy, 0700); err != nil {
			return "", err
		}
		sysInitPath = localCopy
	}
	return sysInitPath, nil
}

func isNetworkDisabled(config *Config) bool {
	return config.Bridge.Iface == disableNetworkBridge
}

func initNetworkController(config *Config) (libnetwork.NetworkController, error) {
	controller, err := libnetwork.New()
	if err != nil {
		return nil, fmt.Errorf("error obtaining controller instance: %v", err)
	}

	// Initialize default driver "null"
	if err := controller.ConfigureNetworkDriver("null", options.Generic{}); err != nil {
		return nil, fmt.Errorf("Error initializing null driver: %v", err)
	}

	// Initialize default network on "null"
	if _, err := controller.NewNetwork("null", "none"); err != nil {
		return nil, fmt.Errorf("Error creating default \"null\" network: %v", err)
	}

	// Initialize default driver "host"
	if err := controller.ConfigureNetworkDriver("host", options.Generic{}); err != nil {
		return nil, fmt.Errorf("Error initializing host driver: %v", err)
	}

	// Initialize default network on "host"
	if _, err := controller.NewNetwork("host", "host"); err != nil {
		return nil, fmt.Errorf("Error creating default \"host\" network: %v", err)
	}

	// Initialize default driver "bridge"
	option := options.Generic{
		"EnableIPForwarding": config.Bridge.EnableIPForward}

	if err := controller.ConfigureNetworkDriver("bridge", options.Generic{netlabel.GenericData: option}); err != nil {
		return nil, fmt.Errorf("Error initializing bridge driver: %v", err)
	}

	netOption := options.Generic{
		"BridgeName":          config.Bridge.Iface,
		"Mtu":                 config.Mtu,
		"EnableIPTables":      config.Bridge.EnableIPTables,
		"EnableIPMasquerade":  config.Bridge.EnableIPMasq,
		"EnableICC":           config.Bridge.InterContainerCommunication,
		"EnableUserlandProxy": config.Bridge.EnableUserlandProxy,
	}

	if config.Bridge.IP != "" {
		ip, bipNet, err := net.ParseCIDR(config.Bridge.IP)
		if err != nil {
			return nil, err
		}

		bipNet.IP = ip
		netOption["AddressIPv4"] = bipNet
	}

	if config.Bridge.FixedCIDR != "" {
		_, fCIDR, err := net.ParseCIDR(config.Bridge.FixedCIDR)
		if err != nil {
			return nil, err
		}

		netOption["FixedCIDR"] = fCIDR
	}

	if config.Bridge.FixedCIDRv6 != "" {
		_, fCIDRv6, err := net.ParseCIDR(config.Bridge.FixedCIDRv6)
		if err != nil {
			return nil, err
		}

		netOption["FixedCIDRv6"] = fCIDRv6
	}

	// --ip processing
	if config.Bridge.DefaultIP != nil {
		netOption["DefaultBindingIP"] = config.Bridge.DefaultIP
	}

	// Initialize default network on "bridge" with the same name
	_, err = controller.NewNetwork("bridge", "bridge",
		libnetwork.NetworkOptionGeneric(options.Generic{
			netlabel.GenericData: netOption,
			netlabel.EnableIPv6:  config.Bridge.EnableIPv6,
		}))
	if err != nil {
		return nil, fmt.Errorf("Error creating default \"bridge\" network: %v", err)
	}

	return controller, nil
}
