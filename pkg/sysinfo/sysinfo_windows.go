// +build windows

package sysinfo

type SysInfo struct {
	// TODO Windows
	// Leaving these in until refactoring done in daemon\create and daemon\info.go
	MemoryLimit            bool
	SwapLimit              bool
	IPv4ForwardingDisabled bool
	AppArmor               bool
}

func New(quiet bool) *SysInfo {
	sysInfo := &SysInfo{}
	return sysInfo
}
