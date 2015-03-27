package kernel

import (
	"fmt"
	"syscall"
	"unsafe"
)

type KernelVersionInfo struct {
	kvi   string
	major int
	minor int
	build int
}

func (k *KernelVersionInfo) String() string {
	return fmt.Sprintf("%d.%d %d (%s)", k.major, k.minor, k.build, k.kvi)
}

func GetKernelVersion() (*KernelVersionInfo, error) {

	KVI := &KernelVersionInfo{"Unknown", 0, 0, 0}

	var h syscall.Handle

	err := syscall.RegOpenKeyEx(syscall.HKEY_LOCAL_MACHINE,
		syscall.StringToUTF16Ptr(`SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion\\`),
		0,
		syscall.KEY_READ,
		&h)
	if err != nil {
		return KVI, err
	}
	defer syscall.RegCloseKey(h)

	var buf [1 << 10]uint16
	var typ uint32
	n := uint32(len(buf) * 2) // api expects array of bytes, not uint16

	err = syscall.RegQueryValueEx(h,
		syscall.StringToUTF16Ptr("BuildLabEx"),
		nil,
		&typ,
		(*byte)(unsafe.Pointer(&buf[0])),
		&n)
	if err != nil {
		return KVI, err
	}

	KVI.kvi = syscall.UTF16ToString(buf[:])

	var dwVersion uint32

	dwVersion, err = syscall.GetVersion()
	if err != nil {
		return KVI, err
	}

	KVI.major = int(dwVersion & 0xFF)
	KVI.minor = int((dwVersion & 0XFF00) >> 8)
	KVI.build = int((dwVersion & 0xFFFF0000) >> 16)

	return KVI, nil

}
