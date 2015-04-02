// +build windows

package argon

import (
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/execdriver"
)

type TtyConsole struct {
}

func NewTtyConsole(processConfig *execdriver.ProcessConfig, pipes *execdriver.Pipes) (*TtyConsole, error) {
	tty := &TtyConsole{}
	return tty, nil
}

func (t *TtyConsole) Resize(h, w int) error {
	log.Debugln("Windows exec ttyconsole: resize not implemented ", h, w)
	// This needs a call into the HCS to set the virtual TTY.
	//return term.SetWinsize(t.MasterPty.Fd(), &term.Winsize{Height: uint16(h), Width: uint16(w)})
	return nil
}

func (t *TtyConsole) Close() error {
	return nil
}
