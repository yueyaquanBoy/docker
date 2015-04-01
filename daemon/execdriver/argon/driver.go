// +build windows

// This is the Windows driver for containers
package argon

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/execdriver"
	"github.com/docker/docker/pkg/hcsshim"
	"gopkg.in/natefinch/npipe.v2"
)

const (
	DriverName = "Windows. 1854"
	Version    = "30-Mar-2015"
)

var (

	// For device redirection passed into the shim layer.
	stdDevices hcsshim.Devices

	inListen, outListen, errListen *npipe.PipeListener
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

	log.Debugln("argon::run c.")
	log.Debugln(" - ID            : ", c.ID)
	log.Debugln(" - RootFs        : ", c.Rootfs)
	log.Debugln(" - InitPath      : ", c.InitPath)
	log.Debugln(" - WorkingDir    : ", c.WorkingDir)
	log.Debugln(" - ConfigPath    : ", c.ConfigPath)
	log.Debugln(" - ProcessLabel  : ", c.ProcessLabel)

	log.Debugln("c.ProcessConfig:")
	log.Debugln(" args      :", c.ProcessConfig.Args)
	log.Debugln(" arguments : ", c.ProcessConfig.Arguments)
	log.Debugln(" entrypoint : ", c.ProcessConfig.Entrypoint)
	log.Debugln(" cmd.args  : ", c.ProcessConfig.Cmd.Args)
	log.Debugln(" cmd.dir  : ", c.ProcessConfig.Cmd.Dir)
	log.Debugln(" cmd.env  : ", c.ProcessConfig.Cmd.Env)
	log.Debugln(" cmd.extrafiles  : ", c.ProcessConfig.Cmd.ExtraFiles)
	log.Debugln(" cmd.path  : ", c.ProcessConfig.Cmd.Path)

	var (
		term execdriver.Terminal
		err  error
	)

	// Make sure the client isn't asking for options which aren't supported
	// by Windows containers.
	err = checkSupportedOptions(c)
	if err != nil {
		return execdriver.ExitStatus{ExitCode: -1}, err
	}

	stdDevices := hcsshim.Devices{}

	// Connect stdin
	if pipes.Stdin != nil {
		stdDevices.StdInPipe = `\\.\pipe\docker\` + c.ID + "-stdin"

		// Listen on the named pipe
		inListen, err = npipe.Listen(stdDevices.StdInPipe)
		if err != nil {
			log.Debugln("Failed to listen on ", stdDevices.StdInPipe, err)
			return execdriver.ExitStatus{ExitCode: -1}, err
		}
		defer inListen.Close()

		// Launch a goroutine to do the accept. We do this so that we can
		// cause an otherwise blocking goroutine to gracefully close when
		// the caller (us) closes the listener
		go stdinAccept(inListen, stdDevices.StdInPipe, pipes.Stdin)

		// TODO c.ProcessConfig.Cmd.Stdin = stdinConn
	}

	// Connect stdout
	// TODO c.ProcessConfig.Cmd.Stdout = stdoutConn
	stdDevices.StdOutPipe = `\\.\pipe\docker\` + c.ID + "-stdout"
	outListen, err = npipe.Listen(stdDevices.StdOutPipe)
	if err != nil {
		log.Debugln("Failed to listen on ", stdDevices.StdOutPipe, err)
		return execdriver.ExitStatus{ExitCode: -1}, err
	}
	defer outListen.Close()
	go stdouterrAccept(outListen, stdDevices.StdOutPipe, pipes.Stdout)

	// Connect stderr
	// TODO c.ProcessConfig.Cmd.Stderr = stderrConn
	stdDevices.StdErrPipe = `\\.\pipe\docker\` + c.ID + "-stderr"
	errListen, err = npipe.Listen(stdDevices.StdErrPipe)
	if err != nil {
		log.Debugln("Failed to listen on ", stdDevices.StdErrPipe, err)
		return execdriver.ExitStatus{ExitCode: -1}, err
	}
	defer errListen.Close()
	go stdouterrAccept(errListen, stdDevices.StdErrPipe, pipes.Stderr)

	if c.ProcessConfig.Tty {
		term, err = NewTtyConsole(&c.ProcessConfig, pipes)
	} else {
		term, err = NewStdConsole(&c.ProcessConfig, pipes)
	}
	c.ProcessConfig.Terminal = term

	// Temporarily create a dummy container with the ID
	configuration := `{` + "\n"
	configuration += ` "SystemType" : "Container",` + "\n"
	configuration += ` "Name" : "john",` + "\n"
	configuration += ` "RootDevicePath" : "C:\\Containers\\test",` + "\n"
	configuration += ` "IsDummy" : true` + "\n"
	configuration += `}` + "\n"

	var emulateTTY uint32

	//if c.ProcessConfig.Tty == true {
	//	emulateTTY = 0x00000001
	//}

	err = hcsshim.CreateComputeSystem(c.ID, configuration)
	if err != nil {
		log.Debugln("Failed to create temporary container ", err)
		return execdriver.ExitStatus{ExitCode: -1}, err
	}

	// Start the container
	log.Debugln("Starting container ", c.ID)
	err = hcsshim.StartComputeSystem(c.ID)
	if err != nil {
		log.Debugln("Failed to start ", err)
		return execdriver.ExitStatus{ExitCode: -1}, err
	}

	// Start the command running in the container.
	var pid uint32
	//pid, err = hcsshim.CreateProcessInComputeSystem(c.ID, "c:\\hcs\\CExecEchoProcess.exe 60", stdDevices)

	// Pretty sure this would get caught earlier, but just in case
	// validate that we have something to run
	if c.ProcessConfig.Entrypoint == "" {
		err = errors.New("No entrypoint specified")
		log.Debugln(err)
		return execdriver.ExitStatus{ExitCode: -1}, err
	}

	// Build the command line of the process
	commandLine := c.ProcessConfig.Entrypoint
	for _, arg := range c.ProcessConfig.Arguments {
		log.Debugln("appending ", arg)
		commandLine += " " + arg
	}
	log.Debugln("commandLine: ", commandLine)

	// Launch the process
	pid, err = hcsshim.CreateProcessInComputeSystem(c.ID,
		commandLine,
		stdDevices,
		emulateTTY)
	if err != nil {
		log.Debugln("CreateProcessInComputeSystem() failed ", err)
		return execdriver.ExitStatus{ExitCode: -1}, err
	}

	//Save the PID as we'll need this in Kill()
	log.Debugln("PID ", pid)
	c.ContainerPid = int(pid)

	// TODO: Still not sure if I need this? JJH 3/20/15
	//term, err = execdriver.NewStdConsole(&c.ProcessConfig, pipes)
	//c.ProcessConfig.Terminal = term

	// Invoke the start callback
	if startCallback != nil {
		startCallback(&c.ProcessConfig, int(pid))
	}

	var exitCode uint32
	exitCode, err = hcsshim.WaitForProcessInComputeSystem(c.ID, pid)
	if err != nil {
		log.Debugln("Failed to WaitForProcessInComputeSystem ", err)
		return execdriver.ExitStatus{ExitCode: -1}, err
	}

	// TODO - What do we do with this exit code????
	log.Debugln("exitcode err", exitCode, err)

	// Stop the container
	log.Debugln("Shutting down container ", c.ID)
	err = hcsshim.ShutdownComputeSystem(c.ID)
	if err != nil {
		// IMPORTANT: Don't fail if fails to change state. It could already have been stopped through kill()
		// Otherwise, the docker daemon will hang in job wait()
		//log.Debugln("Failed to stop ", err)
		//return execdriver.ExitStatus{ExitCode: -1}, err
	}

	return execdriver.ExitStatus{ExitCode: 0}, nil
}

func (d *driver) Terminate(p *execdriver.Command) error {
	log.Debugln("WindowsExec: Terminate() ", p.ID)
	return kill(p.ID, p.ContainerPid)
}

func (d *driver) Kill(p *execdriver.Command, sig int) error {
	log.Debugln("WindowsExec: Kill() ", p.ID, sig)
	return kill(p.ID, p.ContainerPid)
}

func kill(ID string, PID int) error {
	log.Debugln("kill() ", ID)
	log.Debugln("PID", PID)

	// TODO Terminate Process
	err := hcsshim.TerminateProcessInComputeSystem(ID, uint32(PID))
	if err != nil {
		log.Debugln("Failed to Terminate ", err)
		// Ignore errors
		err = nil
	}

	err = hcsshim.ShutdownComputeSystem(ID)
	if err != nil {
		log.Debugln("Failed to kill ", ID)
	}

	return err
}

func (d *driver) Pause(c *execdriver.Command) error {
	return fmt.Errorf("Windows containers cannot be paused")
}

func (d *driver) Unpause(c *execdriver.Command) error {
	return fmt.Errorf("Windows containers cannot be paused")
}

func (i *info) IsRunning() bool {
	var running bool
	running = true // TODO
	return running
}

func (d *driver) Info(id string) execdriver.Info {
	return &info{
		ID:     id,
		driver: d,
	}
}

func (d *driver) Name() string {
	return fmt.Sprintf("%s Date %s", DriverName, Version)
}

func (d *driver) GetPidsForContainer(id string) ([]int, error) {
	return nil, fmt.Errorf("GetPidsForContainer: GetPidsForContainer() not implemented")
}

func (d *driver) Clean(id string) error {
	return nil
}

func (d *driver) Stats(id string) (*execdriver.ResourceStats, error) {
	return nil, fmt.Errorf("Stats() not implemented")
}
