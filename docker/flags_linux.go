package main

import (
	flag "github.com/docker/docker/pkg/mflag"
)

func getDaemonConfDir() string {
	return "/etc/docker"
}

var (
	flWinService = flag.String([]string{"-service"}, "start", "Windows service daemon options")
)
