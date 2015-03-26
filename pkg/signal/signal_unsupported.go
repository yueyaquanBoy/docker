// +build !linux,!darwin,!freebsd,!windows

package signal

import (
	"syscall"
)

var SignalMap = map[string]syscall.Signal{}
