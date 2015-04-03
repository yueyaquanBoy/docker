// +build windows

package argon

import ()

type StdConsole struct {
}

func NewStdConsole(ID string) (*StdConsole, error) {
	return &StdConsole{}, nil
}

func (s *StdConsole) Resize(h, w int) error {
	// we do not need to resize a non tty
	return nil
}

func (s *StdConsole) Close() error {
	// nothing to close here
	return nil
}
