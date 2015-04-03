// +build windows

// Shim for Windows Containers Host Comput Service (HCS)
// This file contains the VHD specific shim function calls.

package hcsshim

import (
	"syscall"
	"unsafe"

	log "github.com/Sirupsen/logrus"
)

var (
	procCreateDiffVhd    = modvmcompute.NewProc("CreateDiffVhd")
	procMountVhd         = modvmcompute.NewProc("MountVhd")
	procDismountVhd      = modvmcompute.NewProc("DismountVhd")
	procGetVhdVolumePath = modvmcompute.NewProc("GetVhdVolumePath")
)

func CreateDiffVhd(newVhdPath, parentVhdPath string) error {
	log.Debugln("hcsshim::CreateDiffVhd")
	log.Debugln("newVhdPath:", newVhdPath)
	log.Debugln("parentVhdPath:", parentVhdPath)

	newVhdPathp, err := syscall.UTF16PtrFromString(newVhdPath)
	if err != nil {
		log.Debugln("Failed conversion of newVhdPath to pointer:", err)
		return err
	}

	parentVhdPathp, err := syscall.UTF16PtrFromString(parentVhdPath)
	if err != nil {
		log.Debugln("Failed conversion of parentVhdPath to pointer:", err)
		return err
	}

	// Call the procedure itself.
	r1, _, _ := procCreateDiffVhd.Call(
		uintptr(unsafe.Pointer(newVhdPathp)),
		uintptr(unsafe.Pointer(parentVhdPathp)))

	if r1 != 0 {
		return syscall.Errno(r1)
	}

	return nil
}

func MountVhd(vhdPath string) (string, error) {
	log.Debugln("hcsshim::MountVhd")
	log.Debugln("vhdPath:", vhdPath)

	vhdPathp, err := syscall.UTF16PtrFromString(vhdPath)
	if err != nil {
		log.Debugln("Failed conversion of vhdPath to pointer:", err)
		return "", err
	}

	var volumePathp [256]uint16
	volumePathp[0] = 0
	// Call the procedure itself.
	r1, _, _ := procMountVhd.Call(
		uintptr(unsafe.Pointer(vhdPathp)),
		uintptr(unsafe.Pointer(&volumePathp)))

	if r1 != 0 {
		return "", syscall.Errno(r1)
	}

	return syscall.UTF16ToString(volumePathp[0:]), nil
}

func DismountVhd(vhdPath string) error {
	log.Debugln("hcsshim::DismountVhd")
	log.Debugln("vhdPath:", vhdPath)

	vhdPathp, err := syscall.UTF16PtrFromString(vhdPath)
	if err != nil {
		log.Debugln("Failed conversion of vhdPath to pointer:", err)
		return err
	}

	r1, _, _ := procDismountVhd.Call(
		uintptr(unsafe.Pointer(vhdPathp)))

	if r1 != 0 {
		return syscall.Errno(r1)
	}

	return nil
}

func GetVhdVolumePath(vhdPath string) (string, error) {
	log.Debugln("hcsshim::GetVhdVolumePath")
	log.Debugln("vhdPath:", vhdPath)

	vhdPathp, err := syscall.UTF16PtrFromString(vhdPath)
	if err != nil {
		log.Debugln("Failed conversion of vhdPath to pointer:", err)
		return "", err
	}

	var volumePathp [256]uint16
	volumePathp[0] = 0
	r1, _, _ := procGetVhdVolumePath.Call(
		uintptr(unsafe.Pointer(vhdPathp)),
		uintptr(unsafe.Pointer(&volumePathp)))

	if r1 != 0 {
		return "", syscall.Errno(r1)
	}

	return syscall.UTF16ToString(volumePathp[0:]), nil
}
