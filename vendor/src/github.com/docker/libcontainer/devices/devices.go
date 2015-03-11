package devices

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/libcontainer/configs"
)

var (
	ErrNotADevice = errors.New("not a device node")
)

// Testing dependencies
var (
	osLstat       = os.Lstat
	ioutilReadDir = ioutil.ReadDir
)

func HostDevices() ([]*configs.Device, error) {
	return getDevices("/dev")
}

func getDevices(path string) ([]*configs.Device, error) {
	files, err := ioutilReadDir(path)
	if err != nil {
		return nil, err
	}
	out := []*configs.Device{}
	for _, f := range files {
		switch {
		case f.IsDir():
			switch f.Name() {
			case "pts", "shm", "fd", "mqueue":
				continue
			default:
				sub, err := getDevices(filepath.Join(path, f.Name()))
				if err != nil {
					return nil, err
				}

				out = append(out, sub...)
				continue
			}
		case f.Name() == "console":
			continue
		}
		device, err := DeviceFromPath(filepath.Join(path, f.Name()), "rwm")
		if err != nil {
			if err == ErrNotADevice {
				continue
			}
			return nil, err
		}
		out = append(out, device)
	}
	return out, nil
}
