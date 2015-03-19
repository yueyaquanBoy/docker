//+build windows

package windowsgraphdriver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/pkg/chrootarchive"
	"github.com/docker/docker/pkg/pshell"
	"github.com/docker/docker/pkg/system"
	"github.com/docker/libcontainer/label"
)

func init() {
	graphdriver.Register("windows", Init)
}

type Driver struct {
	home       string
	sync.Mutex // Protects concurrent modification to active
	active     map[string]int
}

// New returns a new WINDOWS driver.
func Init(home string, options []string) (graphdriver.Driver, error) {
	log.Debugln("WindowsGraphDriver Init() home", home)

	d := &Driver{
		home:   home,
		active: make(map[string]int),
	}

	//return d, nil
	return graphdriver.NaiveDiffDriver(d), nil

}

func (*Driver) String() string {
	log.Debugln("WindowsGraphDriver String()")
	return "windows"
}

func (d *Driver) Status() [][2]string {
	log.Debugln("WindowsGraphDriver Status()")
	return [][2]string{
		{"Windows", "To be confirmed what should be returned by Windows Storage Driver"},
	}
}

// Exists returns true if the given id is registered with
// this driver
func (d *Driver) Exists(id string) bool {
	_, err := system.Lstat(d.dir(id))
	if err == nil {
		log.Debugln("WindowsGraphDriver Exists() - DOES", id, d.dir(id))
	} else {
		log.Debugln("WindowsGraphDriver Exists() - DOES NOT EXIST", id, d.dir(id))
	}
	return err == nil
}

func (d *Driver) Create(id, parent string) error {
	log.Debugln("WindowsGraphDriver Create() id %s, parent %s", id, parent)

	dir := d.dir(id)
	log.Debugln("dir=", dir)
	if err := os.MkdirAll(filepath.Dir(dir), 0700); err != nil {
		return err
	}
	if err := os.Mkdir(dir, 0755); err != nil {
		return err
	}
	opts := []string{"level:s0"}
	if _, mountLabel, err := label.InitLabels(opts); err == nil {
		label.Relabel(dir, mountLabel, "")
	}
	if strings.HasSuffix(parent, "-init") {
		// This is a layer for a container. It should be
		// a mount point into a VHD.
		if err := CreateAndMountVhd(id, dir); err != nil {
			return err
		}
		defer DismountVhd(dir)
	}
	if parent == "" {
		return nil
	}
	parentDir, err := d.Get(parent, "")
	if err != nil {
		return fmt.Errorf("%s: %s", parent, err)
	}

	log.Debugln("Calling chrootarchive.CopyWithTar parentDir", parentDir)
	log.Debugln("Calling chrootarchive.CopyWithTar dir", dir)
	if err := chrootarchive.CopyWithTar(parentDir, dir); err != nil {
		return err
	}
	return nil
}

func (d *Driver) dir(id string) string {
	return filepath.Join(d.home, "dir", filepath.Base(id))
}

// Unmount and remove the dir information
func (d *Driver) Remove(id string) error {
	log.Debugln("WindowsGraphDriver Remove()", id)

	dir := d.dir(id)
	st, err := system.Lstat(dir)
	if err != nil {
		return err
	}

	d.Lock()
	defer d.Unlock()

	if d.active[id] != 0 {
		log.Errorf("Removing active id %s", id)
	}

	if st.Mode()&os.ModeSymlink != 0 {
		// This is a layer for a container. It should be
		// a mount point into a VHD.
		if d.active[id] > 0 {
			if err := DismountVhd(dir); err != nil {
				return err
			}
		}
		if err := os.Remove(dir + ".vhdx"); err != nil {
			return err
		}
	}

	return os.RemoveAll(dir)
}

// Return the rootfs path for the id
// This will mount the dir at it's given path
func (d *Driver) Get(id, mountLabel string) (string, error) {
	log.Debugln("WindowsGraphDriver Get()", id, mountLabel)

	dir := d.dir(id)
	st, err := system.Lstat(dir)
	if err != nil {
		return "", err
	} else if !st.IsDir() {
		return "", fmt.Errorf("%s: not a directory", dir)
	}

	d.Lock()
	defer d.Unlock()

	if st.Mode()&os.ModeSymlink != 0 && d.active[id] == 0 {
		// This is a layer for a container. It should be
		// a mount point into a VHD.
		volumePath, err := MountVhd(dir)
		if err != nil {
			return "", err
		}
		dir = volumePath
	}

	d.active[id]++

	return dir, nil
}

func (d *Driver) Put(id string) error {
	log.Debugln("WindowsGraphDriver Put() %s", id)

	dir := d.dir(id)
	st, err := system.Lstat(dir)
	if err != nil {
		return err
	}

	d.Lock()
	defer d.Unlock()

	d.active[id]--
	if st.Mode()&os.ModeSymlink != 0 && d.active[id] == 0 {
		// This is a layer for a container. It should be
		// a mount point into a VHD.
		if err := DismountVhd(dir); err != nil {
			return err
		}
		delete(d.active, id)
	}

	return nil
}

func (d *Driver) Cleanup() error {
	log.Debugln("WindowsGraphDriver Cleanup()")
	return nil
}

func CreateAndMountVhd(id string, folder string) error {
	// This script will create a VHD as a peer of the given folder, then mount
	// that VHD at the given folder, attempting to clean up if
	// any part of the process fails. NOTE: the indentation must be spaces
	// and not tabs, otherwise the powershell invocation will fail.
	script := `
    $name = "` + id + `"
    $folder = "` + folder + `"
    $path = "$(Split-Path $folder)\$name.vhdx"
    function throwifnull {
        if ($args[0] -eq $null){
            throw
        }
    }
    try {
        $vhd = New-VHD -Path $path -Dynamic -SizeBytes 20gb
        throwifnull $vhd
        $mounted = $vhd | Mount-VHD -Passthru
        throwifnull $mounted
        $disk = $mounted | Get-Disk | Initialize-Disk -PassThru
        throwifnull $disk
        $partition = $disk | New-Partition -UseMaximumSize
        throwifnull $partition
        $volume = $partition | Format-Volume -FileSystem NTFS -NewFileSystemLabel "disk_$name" -Confirm:$false
        throwifnull $volume
        mountvol $folder $volume.path
    }catch{
        if ($mounted){Dismount-VHD $mounted.Path}
        if (Test-Path $path){rm $path}
        throw
    }
    `

	log.Debugln("Attempting to create and mount vhdx named '", id, "'at'", folder, "'")
	_, err := pshell.ExecutePowerShell(script)
	return err
}

func MountVhd(path string) (string, error) {
	// This script will mount the given VHD.
	// NOTE: the indentation must be spaces and not tabs, otherwise the
	// powershell invocation will fail.
	script := `
    $path = "` + path + `.vhdx"
    if(Test-Path $path){
        $mounted = Mount-VHD $path -passthru
        $volume = $mounted | Get-Disk | Get-Partition | Get-Volume
        $volume.Path
    }
    `

	log.Debugln("Attempting to mount VHD '", path, ".vhdx'")
	volumePath, err := pshell.ExecutePowerShell(script)
	return volumePath, err
}

func DismountVhd(path string) error {
	// This script will dismount the given VHD.
	// NOTE: the indentation must be spaces and not tabs, otherwise the
	// powershell invocation will fail.
	script := `
    $path = "` + path + `.vhdx"
    if(Test-Path $path){
        Dismount-VHD $path
    }
    `

	log.Debugln("Attempting to dismount VHD '", path, ".vhdx'")
	_, err := pshell.ExecutePowerShell(script)
	return err
}
