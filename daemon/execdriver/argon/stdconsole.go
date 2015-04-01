// +build windows

package argon

import (
	"github.com/docker/docker/daemon/execdriver"
)

type StdConsole struct {
}

func NewStdConsole(processConfig *execdriver.ProcessConfig, pipes *execdriver.Pipes) (*StdConsole, error) {
	std := &StdConsole{}
	return std, nil
}

func (s *StdConsole) Resize(h, w int) error {
	// we do not need to resize a non tty
	return nil
}

func (s *StdConsole) Close() error {
	// nothing to close here
	return nil
}
