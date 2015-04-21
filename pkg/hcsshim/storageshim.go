// +build windows

// Shim for Windows Containers Host Compute Service (HSC)
// This file contains the storage specific shim function calls.

package hcsshim

import (
	"syscall"
	"unsafe"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/guid"
)

var (
	procAttachFilter  = modvmcompute.NewProc("AttachStorageFilter")
	procDetatchFilter = modvmcompute.NewProc("DetachStorageFilter")
	procInitSandbox   = modvmcompute.NewProc("InitializeStorageSandbox")
	procRemoveFile    = modvmcompute.NewProc("RemoveFileOrReparsePoint")
)

/* To pass into syscall, we need a struct matching the following:
typedef struct _WC_LAYER_DESCRIPTOR {

    //
    // The ID of the layer
    //

    GUID LayerId;

    //
    // Additional flags
    //

    union {
        struct {
            ULONG Reserved : 31;
            ULONG Dirty : 1;    // Created from sandbox as a result of snapshot
        };
        ULONG Value;
    } Flags;

    //
    // Path to the layer root directory, null-terminated
    //

    PCWSTR Path;

} WC_LAYER_DESCRIPTOR, *PWC_LAYER_DESCRIPTOR;
*/
type WC_LAYER_DESCRIPTOR struct {
	LayerId guid.Guid
	Flags   uint32
	Pathp   *uint16
}

func InitializeStorageSandbox(sandboxPath string, parentLayerPaths []string) error {
	log.Debugln("hcsshim::InitializeStorageSandbox")
	log.Debugln("sandboxPath:", sandboxPath)
	log.Debugln("parentLayerPaths:", parentLayerPaths)

	var (
		// Arguments to the call.
		sandboxPathp      *uint16
		layerDescriptorsp *WC_LAYER_DESCRIPTOR

		// Array of descriptors that gets constructed.
		layers []WC_LAYER_DESCRIPTOR

		// Error tracking
		err error
	)

	for i := 0; i < len(parentLayerPaths); i++ {
		// Create a layer descriptor, using the folder path
		// as the source for a GUID LayerId
		g := guid.NewGuid(parentLayerPaths[i])

		p, err := syscall.UTF16PtrFromString(parentLayerPaths[i])
		if err != nil {
			log.Debugln("Failed conversion of parentLayerPath to pointer ", err)
			return err
		}

		layers = append(layers, WC_LAYER_DESCRIPTOR{
			LayerId: *g,
			Flags:   0,
			Pathp:   p,
		})
	}

	sandboxPathp, err = syscall.UTF16PtrFromString(sandboxPath)
	if err != nil {
		log.Debugln("Failed conversion of sandboxPath to pointer ", err)
		return err
	}

	layerDescriptorsp = &(layers[0])

	sandboxPathup := unsafe.Pointer(sandboxPathp)
	layerDescriptorsup := unsafe.Pointer(layerDescriptorsp)

	// Call the procedure itself.
	r1, _, _ := procInitSandbox.Call(
		uintptr(sandboxPathup),
		uintptr(layerDescriptorsup),
		uintptr(len(layers)))
	use(unsafe.Pointer(sandboxPathp))
	use(sandboxPathup)
	use(unsafe.Pointer(layerDescriptorsp))
	use(layerDescriptorsup)

	if r1 != 0 {
		return syscall.Errno(r1)
	}

	return nil
}

func RemoveFileOrReparsePoint(filePath string) error {
	log.Debugln("hcsshim::RemoveFileOrReparsePoint")
	log.Debugln("filePath:", filePath)

	var (
		// Arguments to the call.
		filePathp *uint16

		// Error tracking
		err error
	)

	filePathp, err = syscall.UTF16PtrFromString(filePath)
	if err != nil {
		log.Debugln("Failed conversion of filePath to pointer ", err)
		return err
	}

	filePathup := unsafe.Pointer(filePathp)

	// Call the procedure itself.
	r1, _, _ := procRemoveFile.Call(uintptr(filePathup))
	use(unsafe.Pointer(filePathp))
	use(filePathup)

	if r1 != 0 {
		return syscall.Errno(r1)
	}

	return nil
}
