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
	procLayerExists       = modvmcompute.NewProc("LayerExists")
	procCreateLayer       = modvmcompute.NewProc("CreateLayer")
	procDestroyLayer      = modvmcompute.NewProc("DestroyLayer")
	procActivateLayer     = modvmcompute.NewProc("ActivateLayer")
	procDeactivateLayer   = modvmcompute.NewProc("DeactivateLayer")
	procGetLayerMountPath = modvmcompute.NewProc("GetLayerMountPath")
	procCopyLayer         = modvmcompute.NewProc("CopyLayer")
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
	Flavor  int
	HomeDir string
}

type driverInfo struct {
	Flavor   int
	HomeDirp *uint16
}

func convertInfo(info DriverInfo) (driverInfo, error) {
	homedirp, err := syscall.UTF16PtrFromString(info.HomeDir)
	if err != nil {
		log.Debugln("Failed conversion of home to pointer for driver info: ", err.Error())
		return driverInfo{}, err
	}

	return driverInfo{
		Flavor:   info.Flavor,
		HomeDirp: homedirp,
	}, nil
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

	infop, err := convertInfo(info)
	if err != nil {
		log.Debugln("Failed conversion of driver info ", err)
		return false, err
	}

	var exists bool

	// Call the procedure itself.
	r1, _, _ := procLayerExists.Call(
		uintptr(unsafe.Pointer(&infop)),
		uintptr(unsafe.Pointer(idp)),
		uintptr(unsafe.Pointer(&exists)))
	use(unsafe.Pointer(&infop))
	use(unsafe.Pointer(idp))

	if r1 != 0 {
		return false, syscall.Errno(r1)
	}

	return exists, nil
}

func CreateLayer(info DriverInfo, id, parent string) error {
	log.Debugln("hcsshim::CreateLayer")
	log.Debugln("info.Flavor:", info.Flavor)
	log.Debugln("id:", id)
	log.Debugln("parent:", parent)

	idp, err := syscall.UTF16PtrFromString(id)
	if err != nil {
		log.Debugln("Failed conversion of id to pointer ", err)
		return err
	}

	parentp, err := syscall.UTF16PtrFromString(parent)
	if err != nil {
		log.Debugln("Failed conversion of parent to pointer ", err)
		return err
	}

	infop, err := convertInfo(info)
	if err != nil {
		log.Debugln("Failed conversion of driver info ", err)
		return err
	}

	// Call the procedure itself.
	r1, _, _ := procCreateLayer.Call(
		uintptr(unsafe.Pointer(&infop)),
		uintptr(unsafe.Pointer(idp)),
		uintptr(unsafe.Pointer(parentp)))
	use(unsafe.Pointer(&infop))
	use(unsafe.Pointer(idp))
	use(unsafe.Pointer(parentp))

	if r1 != 0 {
		return syscall.Errno(r1)
	}

	return nil
}

func DestroyLayer(info DriverInfo, id string) error {
	log.Debugln("hcsshim::DestroyLayer")
	log.Debugln("info.Flavor:", info.Flavor)
	log.Debugln("id:", id)

	idp, err := syscall.UTF16PtrFromString(id)
	if err != nil {
		log.Debugln("Failed conversion of id to pointer ", err)
		return err
	}

	infop, err := convertInfo(info)
	if err != nil {
		log.Debugln("Failed conversion of driver info ", err)
		return err
	}

	// Call the procedure itself.
	r1, _, _ := procDestroyLayer.Call(
		uintptr(unsafe.Pointer(&infop)),
		uintptr(unsafe.Pointer(idp)))
	use(unsafe.Pointer(&infop))
	use(unsafe.Pointer(idp))

	if r1 != 0 {
		return syscall.Errno(r1)
	}

	return nil
}

func ActivateLayer(info DriverInfo, id string) error {
	log.Debugln("hcsshim::ActivateLayer")
	log.Debugln("info.Flavor:", info.Flavor)
	log.Debugln("id:", id)

	idp, err := syscall.UTF16PtrFromString(id)
	if err != nil {
		log.Debugln("Failed conversion of id to pointer ", err)
		return err
	}

	infop, err := convertInfo(info)
	if err != nil {
		log.Debugln("Failed conversion of driver info ", err)
		return err
	}

	// Call the procedure itself.
	r1, _, _ := procActivateLayer.Call(
		uintptr(unsafe.Pointer(&infop)),
		uintptr(unsafe.Pointer(idp)))
	use(unsafe.Pointer(&infop))
	use(unsafe.Pointer(idp))

	if r1 != 0 {
		return syscall.Errno(r1)
	}

	return nil
}

func DeactivateLayer(info DriverInfo, id string) error {
	log.Debugln("hcsshim::DeactivateLayer")
	log.Debugln("info.Flavor:", info.Flavor)
	log.Debugln("id:", id)

	idp, err := syscall.UTF16PtrFromString(id)
	if err != nil {
		log.Debugln("Failed conversion of id to pointer ", err)
		return err
	}

	infop, err := convertInfo(info)
	if err != nil {
		log.Debugln("Failed conversion of driver info ", err)
		return err
	}

	// Call the procedure itself.
	r1, _, _ := procDeactivateLayer.Call(
		uintptr(unsafe.Pointer(&infop)),
		uintptr(unsafe.Pointer(idp)))
	use(unsafe.Pointer(&infop))
	use(unsafe.Pointer(idp))

	if r1 != 0 {
		return syscall.Errno(r1)
	}

	return nil
}

func GetLayerMountPath(info DriverInfo, id string) (string, error) {
	log.Debugln("hcsshim::GetLayerMountPath")
	log.Debugln("info.Flavor:", info.Flavor)
	log.Debugln("id:", id)

	idp, err := syscall.UTF16PtrFromString(id)
	if err != nil {
		log.Debugln("Failed conversion of id to pointer ", err)
		return "", err
	}

	infop, err := convertInfo(info)
	if err != nil {
		log.Debugln("Failed conversion of driver info ", err)
		return "", err
	}

	var mountPathp [256]uint16
	mountPathp[0] = 0

	r1, _, _ := procGetLayerMountPath.Call(
		uintptr(unsafe.Pointer(&infop)),
		uintptr(unsafe.Pointer(idp)),
		uintptr(unsafe.Pointer(&mountPathp)))
	use(unsafe.Pointer(&infop))
	use(unsafe.Pointer(idp))

	if r1 != 0 {
		return "", syscall.Errno(r1)
	}

	return syscall.UTF16ToString(mountPathp[0:]), nil
}

func CopyLayer(info DriverInfo, srcId, dstId string) error {
	log.Debugln("hcsshim::CopyLayer")
	log.Debugln("info.Flavor:", info.Flavor)
	log.Debugln("srcId:", srcId)
	log.Debugln("dstId:", dstId)

	srcIdp, err := syscall.UTF16PtrFromString(srcId)
	if err != nil {
		log.Debugln("Failed conversion of srcId to pointer ", err)
		return err
	}

	dstIdp, err := syscall.UTF16PtrFromString(dstId)
	if err != nil {
		log.Debugln("Failed conversion of dstId to pointer ", err)
		return err
	}

	infop, err := convertInfo(info)
	if err != nil {
		log.Debugln("Failed conversion of driver info ", err)
		return err
	}

	// Call the procedure itself.
	r1, _, _ := procCopyLayer.Call(
		uintptr(unsafe.Pointer(&infop)),
		uintptr(unsafe.Pointer(srcIdp)),
		uintptr(unsafe.Pointer(dstIdp)))
	use(unsafe.Pointer(&infop))
	use(unsafe.Pointer(srcIdp))
	use(unsafe.Pointer(dstIdp))

	if r1 != 0 {
		return syscall.Errno(r1)
	}

	return nil
}
