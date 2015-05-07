// +build windows

// Shim for Windows Containers Host Compute Service (HSC)
// This file contains the layer specific shim function calls, used to implement
// the Windows graphdriver functionality.

package hcsshim

import (
	"syscall"
	"unsafe"

	log "github.com/Sirupsen/logrus"
)

var (
	procLayerExists = modvmcompute.NewProc("LayerExists")
)

/* To pass into syscall, we need a struct matching the following:
enum GraphDriverType
{
    DiffDriver,
    FilterDriver
};

struct DriverInfo {
    GraphDriverType Flavor;
    LPCWSTR HomeDir;
};
*/
type DriverInfo struct {
	Flavor   int
	HomeDirp *uint16
}

func LayerExists(info DriverInfo, id string) (bool, error) {
	log.Debugln("hcsshim::LayerExists")
	log.Debugln("info.Flavor:", info.Flavor)
	log.Debugln("id:", id)

	idp, err := syscall.UTF16PtrFromString(id)
	if err != nil {
		log.Debugln("Failed conversion of id to pointer ", err)
		return false, err
	}

	var exists bool

	// Call the procedure itself.
	r1, _, _ := procLayerExists.Call(
		uintptr(unsafe.Pointer(&info)),
		uintptr(unsafe.Pointer(idp)),
		uintptr(unsafe.Pointer(&exists)))
	use(unsafe.Pointer(idp))

	if r1 != 0 {
		return false, syscall.Errno(r1)
	}

	return exists, nil
}
