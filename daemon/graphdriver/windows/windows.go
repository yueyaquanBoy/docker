//+build windows

package windowsgraphdriver

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/pkg/chrootarchive"
	"github.com/docker/libcontainer/label"
)

func init() {
	graphdriver.Register("windows", Init)
}

type Driver struct {
	home string
}

// New returns a new WINDOWS driver.
func Init(home string, options []string) (graphdriver.Driver, error) {

	log.Debugln("WindowsGraphDriver Init() home", home)
	d := &Driver{
		home: home,
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

	_, err := os.Stat(d.dir(id))
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
	if _, err := os.Stat(d.dir(id)); err != nil {
		return err
	}
	return os.RemoveAll(d.dir(id))

}

// Return the rootfs path for the id
// This will mount the dir at it's given path
func (d *Driver) Get(id, mountLabel string) (string, error) {
	log.Debugln("WindowsGraphDriver Get()", id, mountLabel)
	dir := d.dir(id)
	if st, err := os.Stat(dir); err != nil {
		return "", err
	} else if !st.IsDir() {
		return "", fmt.Errorf("%s: not a directory", dir)
	}
	return dir, nil
}

func (d *Driver) Put(id string) error {
	log.Debugln("WindowsGraphDriver Put() %s", id)
	return nil
}

func (d *Driver) Cleanup() error {
	log.Debugln("WindowsGraphDriver Cleanup()")
	return nil
}
