// +build !windows

package chrootarchive

import (
	"github.com/docker/docker/pkg/reexec"
)

func register() {
	reexec.Register("docker-untar", untar)
}
