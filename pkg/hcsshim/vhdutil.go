// +build windows

// Shim for Windows Containers Host Comput Service (HCS)
// This file contains the VHD specific shim function calls.

package hcsshim

import (
	"errors"
	"syscall"
	"unsafe"

	log "github.com/Sirupsen/logrus"
)

const (
	PROCCREATEDIFFVHD = "CreateDiffVhd"
	PROCMOUNTVHD      = "MountVhd"
	PROCDISMOUNTVHD   = "DismountVhd"
	PROCGETVOLUMEPATH = "GetVhdVolumePath"
)

func CreateDiffVhd(newVhdPath, parentVhdPath string) error {
	log.Debugln("hcsshim::CreateDiffVhd")
	log.Debugln("newVhdPath:", newVhdPath)
	log.Debugln("parentVhdPath:", parentVhdPath)

	var (
		// Arguments to the call.
		newVhdPathp, parentVhdPathp *uint16

		// Result values from calling the procedure.
		r1, r2 uintptr

		// The DLL and the procedure in the DLL respectively.
		dll  *syscall.DLL
		proc *syscall.Proc

		// Error tracking.
		err error
	)

	newVhdPathp, err = syscall.UTF16PtrFromString(newVhdPath)
	if err != nil {
		log.Debugln("Failed conversion of newVhdPath to pointer:", err)
		return err
	}

	parentVhdPathp, err = syscall.UTF16PtrFromString(parentVhdPath)
	if err != nil {
		log.Debugln("Failed conversion of parentVhdPath to pointer:", err)
		return err
	}

	// Load the DLL and get the CreateComputeSystem function
	dll, proc, err = loadAndFind(PROCCREATEDIFFVHD)

	// Release once used if we managed to get a handle to it
	if dll != nil {
		defer dll.Release()
	}

	// Call the procedure itself.
	r1, r2, err = proc.Call(
		uintptr(unsafe.Pointer(newVhdPathp)),
		uintptr(unsafe.Pointer(parentVhdPathp)))

	// Check the result codes first
	if r1 != 0 || r2 != 0 {
		// Check for error itself next
		if err != nil {
			if err.Error() != "The operation completed successfully." {
				return errors.New(PROCCREATEDIFFVHD + " failed. " + err.Error())
			}
		}
		log.Debugln("r1 ", r1)
		return errors.New(PROCCREATEDIFFVHD + " failed r1/r2 check")
	}

	return nil
}

func MountVhd(vhdPath string) (string, error) {
	log.Debugln("hcsshim::MountVhd")
	log.Debugln("vhdPath:", vhdPath)

	var (
		// Resulting string.
		volumePath string

		// Arguments to the call.
		vhdPathp    *uint16
		volumePathp [256]uint16

		// Result values from calling the procedure.
		r1, r2 uintptr

		// The DLL and the procedure in the DLL respectively.
		dll  *syscall.DLL
		proc *syscall.Proc

		// Error tracking.
		err error
	)

	volumePathp[0] = 0

	vhdPathp, err = syscall.UTF16PtrFromString(vhdPath)
	if err != nil {
		log.Debugln("Failed conversion of vhdPath to pointer:", err)
		return "", err
	}

	// Load the DLL and get the CreateComputeSystem function
	dll, proc, err = loadAndFind(PROCMOUNTVHD)

	// Release once used if we managed to get a handle to it
	if dll != nil {
		defer dll.Release()
	}

	// Call the procedure itself.
	r1, r2, err = proc.Call(
		uintptr(unsafe.Pointer(vhdPathp)),
		uintptr(unsafe.Pointer(&volumePathp)))

	// Check the result codes first
	if r1 != 0 || r2 != 0 {
		// Check for error itself next
		if err != nil {
			if err.Error() != "The operation completed successfully." {
				return "", errors.New(PROCMOUNTVHD + " failed. " + err.Error())
			}
		}
		log.Debugln("r1 ", r1)
		return "", errors.New(PROCMOUNTVHD + " failed r1/r2 check")
	}

	if volumePathp[0] != 0 {
		volumePath = syscall.UTF16ToString(volumePathp[0:])
	}

	return volumePath, nil
}

func DismountVhd(vhdPath string) error {
	log.Debugln("hcsshim::DismountVhd")
	log.Debugln("vhdPath:", vhdPath)

	var (
		// Arguments to the call.
		vhdPathp *uint16

		// Result values from calling the procedure.
		r1, r2 uintptr

		// The DLL and the procedure in the DLL respectively.
		dll  *syscall.DLL
		proc *syscall.Proc

		// Error tracking.
		err error
	)

	vhdPathp, err = syscall.UTF16PtrFromString(vhdPath)
	if err != nil {
		log.Debugln("Failed conversion of vhdPath to pointer:", err)
		return err
	}

	// Load the DLL and get the CreateComputeSystem function
	dll, proc, err = loadAndFind(PROCDISMOUNTVHD)

	// Release once used if we managed to get a handle to it
	if dll != nil {
		defer dll.Release()
	}

	// Call the procedure itself.
	r1, r2, err = proc.Call(
		uintptr(unsafe.Pointer(vhdPathp)))

	// Check the result codes first
	if r1 != 0 || r2 != 0 {
		// Check for error itself next
		if err != nil {
			if err.Error() != "The operation completed successfully." {
				return errors.New(PROCDISMOUNTVHD + " failed. " + err.Error())
			}
		}
		log.Debugln("r1 ", r1)
		return errors.New(PROCDISMOUNTVHD + " failed r1/r2 check")
	}

	return nil
}

func GetVhdVolumePath(vhdPath string) (string, error) {
	log.Debugln("hcsshim::GetVhdVolumePath")
	log.Debugln("vhdPath:", vhdPath)

	var (
		// Resulting string.
		volumePath string

		// Arguments to the call.
		vhdPathp    *uint16
		volumePathp [256]uint16

		// Result values from calling the procedure.
		r1, r2 uintptr

		// The DLL and the procedure in the DLL respectively.
		dll  *syscall.DLL
		proc *syscall.Proc

		// Error tracking.
		err error
	)

	volumePathp[0] = 0

	vhdPathp, err = syscall.UTF16PtrFromString(vhdPath)
	if err != nil {
		log.Debugln("Failed conversion of vhdPath to pointer:", err)
		return "", err
	}

	// Load the DLL and get the CreateComputeSystem function
	dll, proc, err = loadAndFind(PROCGETVOLUMEPATH)

	// Release once used if we managed to get a handle to it
	if dll != nil {
		defer dll.Release()
	}

	// Call the procedure itself.
	r1, r2, err = proc.Call(
		uintptr(unsafe.Pointer(vhdPathp)),
		uintptr(unsafe.Pointer(&volumePathp)))

	// Check the result codes first
	if r1 != 0 || r2 != 0 {
		// Check for error itself next
		if err != nil {
			if err.Error() != "The operation completed successfully." {
				return "", errors.New(PROCGETVOLUMEPATH + " failed. " + err.Error())
			}
		}
		log.Debugln("r1 ", r1)
		return "", errors.New(PROCGETVOLUMEPATH + " failed r1/r2 check")
	}

	if volumePathp[0] != 0 {
		volumePath = syscall.UTF16ToString(volumePathp[0:])
	}

	return volumePath, nil
}
