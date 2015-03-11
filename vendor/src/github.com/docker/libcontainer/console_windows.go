package libcontainer

import (
	"os"
)

// newConsole returns an initalized console that can be used within a container by copying bytes
// from the master side to the slave that is attached as the tty for the container's init process.
func newConsole(uid, gid int) (Console, error) {
	return nil, nil
}

// linuxConsole is a linux psuedo TTY for use within a container.
type linuxConsole struct {
	master    *os.File
	slavePath string
}

func (c *linuxConsole) Fd() uintptr {
	return 0
}

func (c *linuxConsole) Path() string {
	return ""
}

func (c *linuxConsole) Read(b []byte) (int, error) {
	return 0, nil
}

func (c *linuxConsole) Write(b []byte) (int, error) {
	return 0, nil
}

func (c *linuxConsole) Close() error {
	return nil
}
