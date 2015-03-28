package main

import (
	"path/filepath"

	"github.com/docker/docker/pkg/homedir"
	flag "github.com/docker/docker/pkg/mflag"
)

// TODO BUGBUG Windows. To run as a service, we don't want this is in
// a user home directory. We should be putting this in something like
// %PROGRAMDATA%\Docker
func getDaemonConfDir() string {
	return filepath.Join(homedir.Get(), ".docker")
}

var (
	flWinService = flag.String([]string{"-service"}, "start", "Windows service daemon options")
)
