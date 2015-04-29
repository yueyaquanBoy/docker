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
	procCreateBaseVhd    = modvmcompute.NewProc("CreateBaseVhd")
	procCreateDiffVhd    = modvmcompute.NewProc("CreateDiffVhd")
	procFormatVhd        = modvmcompute.NewProc("FormatVhd")
	procMountVhd         = modvmcompute.NewProc("MountVhd")
	procDismountVhd      = modvmcompute.NewProc("DismountVhd")
	procGetVhdVolumePath = modvmcompute.NewProc("GetVhdVolumePath")
)

func CreateBaseVhd(newVhdPath string, newSize uint64) error {
	log.Debugln("hcsshim::CreateBaseVhd")
	log.Debugln("newVhdPath:", newVhdPath)

	newVhdPathp, err := syscall.UTF16PtrFromString(newVhdPath)
	if err != nil {
		log.Debugln("Failed conversion of newVhdPath to pointer:", err)
		return err
	}

	// Call the procedure itself.
	r1, _, _ := procCreateBaseVhd.Call(
		uintptr(unsafe.Pointer(newVhdPathp)),
		uintptr(newSize))

	use(unsafe.Pointer(newVhdPathp))

	if r1 != 0 {
		return syscall.Errno(r1)
	}

	return nil
}

func FormatVhd(vhdPath string) error {
	log.Debugln("hcsshim::FormatVhd")
	log.Debugln("vhdPath:", vhdPath)

	vhdPathp, err := syscall.UTF16PtrFromString(vhdPath)
	if err != nil {
		log.Debugln("Failed conversion of vhdPath to pointer:", err)
		return err
	}

	// Call the procedure itself.
	r1, _, _ := procFormatVhd.Call(
		uintptr(unsafe.Pointer(vhdPathp)))

	use(unsafe.Pointer(vhdPathp))

	if r1 != 0 {
		return syscall.Errno(r1)
	}

	return nil
}

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

	use(unsafe.Pointer(newVhdPathp))
	use(unsafe.Pointer(parentVhdPathp))

	if r1 != 0 {
		return syscall.Errno(r1)
	}

	return nil
}

func MountVhd(vhdPath string) error {
	log.Debugln("hcsshim::MountVhd")
	log.Debugln("vhdPath:", vhdPath)

	vhdPathp, err := syscall.UTF16PtrFromString(vhdPath)
	if err != nil {
		log.Debugln("Failed conversion of vhdPath to pointer:", err)
		return err
	}

	// Call the procedure itself.
	r1, _, _ := procMountVhd.Call(
		uintptr(unsafe.Pointer(vhdPathp)))

	use(unsafe.Pointer(vhdPathp))

	if r1 != 0 {
		return syscall.Errno(r1)
	}

	return nil
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

	use(unsafe.Pointer(vhdPathp))

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

	use(unsafe.Pointer(vhdPathp))
	use(unsafe.Pointer(&volumePathp))

	if r1 != 0 {
		return "", syscall.Errno(r1)
	}

	return syscall.UTF16ToString(volumePathp[0:]), nil
}
