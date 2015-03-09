// +build windows

// This is the argon driver
package argon

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/execdriver"
	"github.com/docker/docker/pkg/term"
)

const (
	DriverName = "Windows Containers"
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

func (d *driver) Run(c *execdriver.Command, pipes *execdriver.Pipes, startCallback execdriver.StartCallback) (execdriver.ExitStatus, error) {
	// Partial implementation. Just runs notepad for now.
	// TODO Windows
	log.Debugln("windowsexec::run c.")
	log.Debugln(" - ID            : ", c.ID)
	log.Debugln(" - RootFs        : ", c.Rootfs)
	log.Debugln(" - ReadonlyRootfs: ", c.ReadonlyRootfs)
	log.Debugln(" - InitPath      : ", c.InitPath)
	log.Debugln(" - WorkingDir    : ", c.WorkingDir)
	log.Debugln(" - ConfigPath    : ", c.ConfigPath)
	log.Debugln(" - ProcessLabel  : ", c.ProcessLabel)

	// cmd.exe is a good a thing as any to fire up for now
	params := []string{
		"notepad",
		"c:/users/jhoward/documents/a.a",
	}

	var (
		name = params[0]
		arg  = params[1:]
	)
	log.Debugln("name", name)
	log.Debugln("arg", arg)

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

type TtyConsole struct {
	MasterPty *os.File
}

func NewTtyConsole(processConfig *execdriver.ProcessConfig, pipes *execdriver.Pipes) (*TtyConsole, error) {
	return nil, nil
}

func (t *TtyConsole) Master() *os.File {
	return t.MasterPty
}

func (t *TtyConsole) Resize(h, w int) error {
	return term.SetWinsize(t.MasterPty.Fd(), &term.Winsize{Height: uint16(h), Width: uint16(w)})
}

func (t *TtyConsole) AttachPipes(command *exec.Cmd, pipes *execdriver.Pipes) error {
	go func() {
		if wb, ok := pipes.Stdout.(interface {
			CloseWriters() error
		}); ok {
			defer wb.CloseWriters()
		}

		io.Copy(pipes.Stdout, t.MasterPty)
	}()

	if pipes.Stdin != nil {
		go func() {
			io.Copy(t.MasterPty, pipes.Stdin)

			pipes.Stdin.Close()
		}()
	}

	return nil
}

func (t *TtyConsole) Close() error {
	return t.MasterPty.Close()
}
