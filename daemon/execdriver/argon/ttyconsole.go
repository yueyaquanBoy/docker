// +build windows

package argon

import (
	"github.com/docker/docker/pkg/hcsshim"
)

type TtyConsole struct {
	ID string
}

func NewTtyConsole(ID string) (*TtyConsole, error) {
	tty := &TtyConsole{ID: ID}
	return tty, nil
}

func (t *TtyConsole) Resize(ID string, h, w int) error {
	// We need to tell the virtual TTY via HCS that the client has resized.
	return hcsshim.ResizeTTY(ID, h, w)
}

func (t *TtyConsole) Close() error {
	return nil
}
