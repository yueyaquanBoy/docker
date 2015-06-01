package chrootarchive

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/docker/docker/pkg/reexec"
)

func init() {
	register()
	// TODO Windows: Fix applyLayer and call inline rather than reexec.
	reexec.Register("docker-applyLayer", applyLayer)
}

func fatal(err error) {
	fmt.Fprint(os.Stderr, err)
	os.Exit(1)
}

// flush consumes all the bytes from the reader discarding
// any errors
func flush(r io.Reader) {
	io.Copy(ioutil.Discard, r)
}
