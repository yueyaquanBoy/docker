// +build windows

// This is the Windows driver for containers
package argon

import (
	"errors"
	"fmt"
	"os/exec"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/execdriver"
	"github.com/docker/docker/pkg/hcsshim"
	_ "gopkg.in/natefinch/npipe.v2"
)

const (
	DriverName = "1854"
	Version    = "0.1"
)

type driver struct {
	root     string
	initPath string
}

type info struct {
	ID     string
	driver *driver
}

func NewDriver(root, initPath string) (*driver, error) {

	return &driver{
		root:     root,
		initPath: initPath,
	}, nil
}

func (d *driver) Exec(c *execdriver.Command, processConfig *execdriver.ProcessConfig, pipes *execdriver.Pipes, startCallback execdriver.StartCallback) (int, error) {
	return 0, nil
}

func checkSupportedOptions(c *execdriver.Command) error {
	// Windows doesn't support read-only root filesystem
	if c.ReadonlyRootfs {
		return errors.New("Windows does not support the read-only root filesystem option")
	}

	// Windows doesn't support username
	if c.ProcessConfig.User != "" {
		return errors.New("Windows does not support the username option")
	}

	// Windows doesn't support custom lxc options
	if c.LxcConfig != nil {
		return errors.New("Windows does not support lxc options")
	}

	// Windows doesn't support ulimit
	if c.Resources.Rlimits != nil {
		return errors.New("Windows does not support ulimit options")
	}

	return nil

	// NOTSURE:
	//--add-host=[]              Add a custom host-to-IP mapping (host:ip)
	//-c, --cpu-shares=0         CPU shares (relative weight)
	//--cidfile=                 Write the container ID to the file
	//--cpuset-cpus=             CPUs in which to allow execution (0-3, 0,1)
	//--dns=[]                   Set custom DNS servers
	//--dns-search=[]            Set custom DNS search domains
	//-e, --env=[]               Set environment variables
	//--entrypoint=              Overwrite the default ENTRYPOINT of the image
	//--env-file=[]              Read in a file of environment variables
	//--expose=[]                Expose a port or a range of ports
	//-i, --interactive=false    Keep STDIN open even if not attached
	//-m, --memory=              Memory limit
	//--mac-address=             Container MAC address (e.g. 92:d0:c6:0a:29:33)
	//--memory-swap=             Total memory (memory + swap), '-1' to disable swap
	//--name=                    Assign a name to the container
	//--net=bridge               Set the Network mode for the container
	//-P, --publish-all=false    Publish all exposed ports to random ports
	//-p, --publish=[]           Publish a container's port(s) to the host

	// TODO (Block)
	//--cap-add=[]               Add Linux capabilities
	//--cap-drop=[]              Drop Linux capabilities
	//--device=[]                Add a host device to the container
	//-h, --hostname=            Container host name
	//--ipc=                     IPC namespace to use
	//--link=[]                  Add link to another container
	//DONE --lxc-conf=[]              Add custom lxc options
	//--pid=                     PID namespace to use
	//--privileged=false         Give extended privileges to this container
	//--restart=no               Restart policy to apply when a container exits
	//DONE --read-only=false          Mount the container's root filesystem as read only
	//DONE -u, --user=                Username or UID (format: <name|uid>[:<group|gid>])
	//DONE --ulimit=[]                Ulimit options

	// Allow
	//-d, --detach=false         Run container in background and print container ID

	//--security-opt=[]          Security Options
	//--sig-proxy=true           Proxy received signals to the process
	//-t, --tty=false            Allocate a pseudo-TTY

	//-v, --volume=[]            Bind mount a volume
	//--volumes-from=[]          Mount volumes from the specified container(s)
	//-w, --workdir=             Working directory inside the container

}

func (d *driver) Run(c *execdriver.Command, pipes *execdriver.Pipes, startCallback execdriver.StartCallback) (execdriver.ExitStatus, error) {

	// JJH Wherever pipes is set on call in, need to actually
	// create 3 named pipes first.

	var (
		term execdriver.Terminal
		//devices hcsshim.Devices
	)

	hcsshim.ChangeState("12345", hcsshim.Start)

	term, err := execdriver.NewStdConsole(&c.ProcessConfig, pipes)
	c.ProcessConfig.Terminal = term

	// Keep this block of code -- it at least compiles.
	//l := npipe.PipeConn{}
	//pipes.Stdout = &l
	//pipes.Stdin = &l

	// Make sure the client isn't asking for options which aren't supported
	// by Windows containers.
	err = checkSupportedOptions(c)
	if err != nil {
		return execdriver.ExitStatus{ExitCode: -1}, err
	}

	// Partial implementation. Just runs notepad for now.
	// TODO Windows
	log.Debugln("windowsexec::run c.")
	log.Debugln(" - ID            : ", c.ID)
	log.Debugln(" - RootFs        : ", c.Rootfs)
	log.Debugln(" - InitPath      : ", c.InitPath)
	log.Debugln(" - WorkingDir    : ", c.WorkingDir)
	log.Debugln(" - ConfigPath    : ", c.ConfigPath)
	log.Debugln(" - ProcessLabel  : ", c.ProcessLabel)

	// cmd.exe is a good a thing as any to fire up for now
	//	params := []string{
	//		"notepad",
	//		"c:/users/jhoward/documents/a.a",
	//	}

	params := []string{
		"notepad",
	}

	// docker run -a stdin -a stdout -a stderr -i cirros echo hello
	//params := []string{
	//	"c:/windows/system32/cmd.exe",
	//}

	var (
		name = params[0]
		arg  = params[1:]
	)
	log.Debugln("name", name)
	log.Debugln("arg", arg)

	// OK, at this point, need to create 3 pipes

	aname, err := exec.LookPath(name)
	if err != nil {
		log.Debugln("Error returned from exec.LookPath", name, err)
		aname = name
	}
	c.ProcessConfig.Path = aname
	c.ProcessConfig.Args = append([]string{name}, arg...)

	log.Debugln("windowsexec::run c.ProcessConfig")
	log.Debugln(" - Path          : ", c.ProcessConfig.Path)
	log.Debugln(" - Args          : ", c.ProcessConfig.Args)
	log.Debugln(" - Env           : ", c.ProcessConfig.Env)
	log.Debugln(" - Dir           : ", c.ProcessConfig.Dir)

	if err := c.ProcessConfig.Start(); err != nil {
		log.Debugln("ProcessConfig.Start() failed ", err)
		return execdriver.ExitStatus{ExitCode: -1}, err
	}

	var (
		waitErr  error
		waitLock = make(chan struct{})
	)

	go func() {
		if err := c.ProcessConfig.Wait(); err != nil {
			if _, ok := err.(*exec.ExitError); !ok { // Do not propagate the error if it's simply a status code != 0
				waitErr = err
			}
		}
		close(waitLock)
	}()

	// Poll for RUNNING status
	pid, err := d.waitForStart(c, waitLock)
	if err != nil {
		if c.ProcessConfig.Process != nil {
			c.ProcessConfig.Process.Kill()
			c.ProcessConfig.Wait()
		}
		return execdriver.ExitStatus{ExitCode: -1}, err
	}

	// JJH For now, get the pid out of the ProcessConfig object.
	c.ContainerPid = c.ProcessConfig.Process.Pid
	pid = c.ProcessConfig.Process.Pid

	log.Debugln("PID of 'container' is ", c.ContainerPid)

	if startCallback != nil {
		startCallback(&c.ProcessConfig, pid)
	}

	<-waitLock

	return execdriver.ExitStatus{ExitCode: 0}, nil
}

// wait for the process to start and return the pid for the process
func (d *driver) waitForStart(c *execdriver.Command, waitLock chan struct{}) (int, error) {
	var (
		err error
		//output []byte
	)
	// We wait for the container to be fully running.
	// Timeout after 5 seconds. In case of broken pipe, just retry.
	// Note: The container can run and finish correctly before
	// the end of this loop
	for now := time.Now(); time.Since(now) < 5*time.Second; {
		select {
		case <-waitLock:
			// If the process dies while waiting for it, just return
			return -1, nil
		default:
		}

		//output, err = d.getInfo(c.ID)
		if err == nil {
			//info, err := parseLxcInfo(string(output))
			//if err != nil {
			//	return -1, err
			//}
			//if info.Running {
			// Windows - for now, pretend it is running
			return 56789, nil
			//}
		}
		time.Sleep(50 * time.Millisecond)
	}
	return -1, execdriver.ErrNotRunning
}

func (d *driver) Kill(p *execdriver.Command, sig int) error {

	// JJH Hacked together. Just kill the PID we spawned
	log.Debugln("Kill() pid=", p.ProcessConfig.Process.Pid)
	log.Debugln("Kill() sig=", sig)

	err := p.ProcessConfig.Process.Kill()
	return err
}

func (d *driver) Pause(c *execdriver.Command) error {
	return fmt.Errorf("windowsexec: Pause() not implemented but would pause PID %d", c.ProcessConfig.Process.Pid)
}

func (d *driver) Unpause(c *execdriver.Command) error {
	return fmt.Errorf("windowsexec: Pause() not implemented but would Unpause PID %d", c.ProcessConfig.Process.Pid)
}

func (d *driver) Terminate(p *execdriver.Command) error {
	//return Kill(p,9)
	return fmt.Errorf("windowsexec: Terminate() not implemented")
}

func (i *info) IsRunning() bool {
	var running bool
	running = true
	return running
}

func (d *driver) Info(id string) execdriver.Info {
	return &info{
		ID:     id,
		driver: d,
	}
}

func (d *driver) Name() string {
	return fmt.Sprintf("%s-%s", DriverName, Version)
}

func (d *driver) GetPidsForContainer(id string) ([]int, error) {
	return nil, fmt.Errorf("GetPidsForContainer: GetPidsForContainer() not implemented")
}

func (d *driver) Clean(id string) error {
	return nil
}

func (d *driver) Stats(id string) (*execdriver.ResourceStats, error) {
	return nil, fmt.Errorf("windowsexec: Stats() not implemented")
}
