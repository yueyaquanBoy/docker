//+build windows

package windows

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/chrootarchive"
	"github.com/docker/docker/pkg/hcsshim"
	"github.com/docker/docker/pkg/system"
)

func init() {
	graphdriver.Register("windows", InitDiff)
	graphdriver.Register("windowsfilter", InitFilter)
}

type driverType int

const (
	diffDriver driverType = iota
	filterDriver
)

type WindowsGraphDriver struct {
	home       string
	sync.Mutex // Protects concurrent modification to active
	active     map[string]int
	flavor     driverType
}

// New returns a new WINDOWS Diff Disk driver.
func InitDiff(home string, options []string) (graphdriver.Driver, error) {
	log.Debugln("WindowsGraphDriver InitDiff() home", home)

	d := &WindowsGraphDriver{
		home:   home,
		active: make(map[string]int),
		flavor: diffDriver,
	}

	//return d, nil
	return d, nil

}

// New returns a new WINDOWS Storage Filter driver.
func InitFilter(home string, options []string) (graphdriver.Driver, error) {
	log.Debugln("WindowsGraphDriver InitFilter() home", home)

	d := &WindowsGraphDriver{
		home:   home,
		active: make(map[string]int),
		flavor: filterDriver,
	}

	//return d, nil
	return d, nil

}

func (d *WindowsGraphDriver) String() string {
	log.Debugln("WindowsGraphDriver String()")
	switch d.flavor {
	case diffDriver:
		return "windows"
	case filterDriver:
		return "windowsfilter"
	default:
		panic("unsupported driver type")
	}
}

func (d *WindowsGraphDriver) Status() [][2]string {
	log.Debugln("WindowsGraphDriver Status()")
	return [][2]string{
		{"Windows", "To be confirmed what should be returned by Windows Storage Driver"},
	}
}

// Exists returns true if the given id is registered with
// this driver
func (d *WindowsGraphDriver) Exists(id string) bool {
	_, err := system.Lstat(d.dir(id))
	if err == nil {
		log.Debugln("WindowsGraphDriver Exists() - DOES", id, d.dir(id))
	} else {
		log.Debugln("WindowsGraphDriver Exists() - DOES NOT EXIST", id, d.dir(id))
	}
	return err == nil
}

func (d *WindowsGraphDriver) Create(id, parent string) error {
	log.Debugln("WindowsGraphDriver Create() id:", id, ", parent:", parent)

	dir := d.dir(id)
	log.Debugln("dir=", dir)
	if err := system.MkdirAll(filepath.Dir(dir), 0700); err != nil {
		return err
	}
	if err := os.Mkdir(dir, 0755); err != nil {
		return err
	}

	if d.flavor == diffDriver {
		if parent == "" {
			// This is a base layer, so create a new VHD.
			return createBaseVhd(id, dir)
		} else {
			// This is an intermediate layer, so create a diff-VHD from
			// the parent.
			parentDir := d.dir(parent)
			log.Debugln("parentDir=", parentDir)
			return createDiffVhd(id, dir, parentDir)
		}
	}

	return nil
}

func (d *WindowsGraphDriver) dir(id string) string {
	return filepath.Join(d.home, filepath.Base(id))
}

// Unmount and remove the dir information
func (d *WindowsGraphDriver) Remove(id string) error {
	log.Debugln("WindowsGraphDriver Remove()", id)

	dir := d.dir(id)

	d.Lock()
	defer d.Unlock()

	if d.active[id] != 0 {
		log.Warnf("Removing active id %s", id)
	}

	if d.flavor == diffDriver {
		if d.active[id] > 0 {
			if err := dismountVhd(dir); err != nil {
				return err
			}
		}
		if err := os.Remove(dir + ".vhdx"); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return os.RemoveAll(dir)
}

// Return the rootfs path for the id
// This will mount the dir at it's given path
func (d *WindowsGraphDriver) Get(id, mountLabel string) (string, error) {
	log.Debugln("WindowsGraphDriver Get() id:", id, ", mountLabel:", mountLabel)

	dir := d.dir(id)
	st, err := system.Lstat(dir)
	if err != nil {
		return "", err
	} else if !st.IsDir() {
		return "", fmt.Errorf("%s: not a directory", dir)
	}

	d.Lock()
	defer d.Unlock()

	if d.flavor == diffDriver {
		if d.active[id] == 0 {
			dir, err = mountVhd(dir)
		} else {
			dir, err = getMountedVolumePath(dir)
		}
		if err != nil {
			return "", err
		}
	}

	d.active[id]++

	return dir, nil
}

func (d *WindowsGraphDriver) Put(id string) error {
	log.Debugln("WindowsGraphDriver Put() id:", id)

	dir := d.dir(id)

	d.Lock()
	defer d.Unlock()

	if d.active[id] > 1 {
		d.active[id]--
	} else if d.active[id] == 1 {
		if d.flavor == diffDriver {
			if err := dismountVhd(dir); err != nil {
				return err
			}
		}
		delete(d.active, id)
	}

	return nil
}

func (d *WindowsGraphDriver) Cleanup() error {
	log.Debugln("WindowsGraphDriver Cleanup()")
	return nil
}

// Diff produces an archive of the changes between the specified
// layer and its parent layer which may be "".
func (d *WindowsGraphDriver) Diff(id, parent string) (arch archive.Archive, err error) {
	// Always include the diff disk for the given layer.
	if d.flavor == diffDriver {
		diffFiles := []string{id + ".vhdx"}
		prevLayer := d.dir(parent)
		curParent, err := getParentVhdPath(d.dir(id) + ".vhdx")
		if err != nil {
			return nil, err
		}
		for strings.EqualFold((prevLayer + ".vhdx"), curParent) {
			log.Debugf("Parent %s does not match parent %s.", (prevLayer + ".vhdx"), curParent)
			// Add that diff to the list of files.
			_, diffFile := filepath.Split(curParent)
			diffFiles = append(diffFiles, diffFile)
			curParent, err = getParentVhdPath(curParent)
			if err != nil {
				return nil, err
			}
		}

		opts := &archive.TarOptions{
			IncludeFiles: diffFiles,
		}

		arch, err = archive.TarWithOptions(d.home, opts)
		if err != nil {
			return nil, err
		}
		return arch, nil
	} else {
		return nil, fmt.Errorf("Windows Filter Driver: Not Implemented")
	}
}

// Changes produces a list of changes between the specified layer
// and its parent layer. If parent is "", then all changes will be ADD changes.
func (d *WindowsGraphDriver) Changes(id, parent string) ([]archive.Change, error) {
	layerFs, err := d.Get(id, "")
	if err != nil {
		return nil, err
	}
	defer d.Put(id)

	parentFs := ""

	if parent != "" {
		parentFs, err = d.Get(parent, "")
		if err != nil {
			return nil, err
		}
		defer d.Put(parent)
	}

	return archive.ChangesDirs(layerFs, parentFs)
}

// ApplyDiff extracts the changeset from the given diff into the
// layer with the specified id and parent, returning the size of the
// new layer in bytes.
func (d *WindowsGraphDriver) ApplyDiff(id, parent string, diff archive.ArchiveReader) (size int64, err error) {
	start := time.Now().UTC()
	log.Debugf("Start untar layer")

	destination := d.dir(id)
	if d.flavor == diffDriver {
		destination = filepath.Dir(destination)
	}

	if size, err = chrootarchive.ApplyLayer(destination, diff); err != nil {
		return
	}
	log.Debugf("Untar time: %vs", time.Now().UTC().Sub(start).Seconds())

	return
}

// DiffSize calculates the changes between the specified layer
// and its parent and returns the size in bytes of the changes
// relative to its base filesystem directory.
func (d *WindowsGraphDriver) DiffSize(id, parent string) (size int64, err error) {
	changes, err := d.Changes(id, parent)
	if err != nil {
		return
	}

	layerFs, err := d.Get(id, "")
	if err != nil {
		return
	}
	defer d.Put(id)

	return archive.ChangesSize(layerFs, changes), nil
}

func (d *WindowsGraphDriver) CopyDiff(sourceId, id string) error {
	d.Lock()
	defer d.Unlock()

	if d.flavor == diffDriver {
		if d.active[sourceId] != 0 {
			log.Warnf("Committing active id %s", id)
			err := dismountVhd(d.dir(sourceId))
			if err != nil {
				return err
			}
			defer func() {
				_, err := mountVhd(d.dir(sourceId))
				if err != nil {
					log.Warnf("Failed to remount VHD: %s", err)
				}
			}()
		}

		if err := os.Mkdir(d.dir(id), 0755); err != nil {
			return err
		}

		return copyVhd(d.dir(sourceId), d.dir(id))
	} else {
		return fmt.Errorf("Windows Filter Driver: Not Implemented")
	}
}

func createBaseVhd(id string, folder string) error {
	newVhdPath := filepath.Join(filepath.Dir(folder), id) + ".vhdx"

	if err := hcsshim.CreateBaseVhd(newVhdPath, 20); err != nil {
		return err
	}

	return hcsshim.FormatVhd(newVhdPath)
}

func getParentVhdPath(vhdPath string) (string, error) {
	return hcsshim.GetVhdParentPath(vhdPath)
}

func createDiffVhd(id string, folder string, parent string) error {
	newVhdPath := filepath.Join(filepath.Dir(folder), id) + ".vhdx"
	parentVhdPath := parent + ".vhdx"

	return hcsshim.CreateDiffVhd(newVhdPath, parentVhdPath)
}

func mountVhd(path string) (string, error) {
	vhdPath := path + ".vhdx"

	if err := hcsshim.MountVhd(vhdPath); err != nil {
		return "", err
	}

	volPath, err := hcsshim.GetVhdVolumePath(vhdPath)
	if err != nil {
		if err2 := hcsshim.DismountVhd(vhdPath); err2 != nil {
			log.Errorf("Failed to dismount disk '%s': %s", vhdPath, err2.Error())
		}
		return "", err
	}

	return volPath, nil
}

func dismountVhd(path string) error {
	vhdPath := path + ".vhdx"

	return hcsshim.DismountVhd(vhdPath)
}

func getMountedVolumePath(path string) (string, error) {
	vhdPath := path + ".vhdx"

	return hcsshim.GetVhdVolumePath(vhdPath)
}

func copyVhd(src, dst string) error {
	srcPath := src + ".vhdx"
	dstPath := dst + ".vhdx"

	return chrootarchive.CopyFileWithTar(srcPath, dstPath)
}
