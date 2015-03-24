// +build windows

// Shim for Windows Containers Host Compute Service (HSC)

package hcsshim

import (
	"errors"
	"syscall"
	"unsafe"

	log "github.com/Sirupsen/logrus"
)

const (
	SHIMDLL = "vmcompute.dll"

	PROCCREATE                       = "CreateComputeSystem"
	PROCSTART                        = "StartComputeSystem"
	PROCTERMINATE                    = "TerminateComputeSystem"
	PROCRUNANDWAIT                   = "ExecuteInComputeSystem"
	PROCCREATEPROCESSINCOMPUTESYSTEM = "CreateProcessInComputeSystem"
)

var (
	Start = 1
	Stop  = 2
)

// The redirection devices as passed in from callers
type Devices struct {
	StdInPipe  string
	StdOutPipe string
	StdErrPipe string
}

// The redirection devices as passed used internally
type deviceInt struct {
	stdinpipe  *uint16
	stdoutpipe *uint16
	stderrpipe *uint16
}

// Configuration will be JSON such as:

//configuration := `{` + "\n"
//configuration += ` "SystemType" : "Container",` + "\n"
//configuration += ` "Name" : "test2",` + "\n"
//configuration += ` "RootDevicePath" : "C:\\Containers\\test",` + "\n"
//configuration += ` "IsDummy" : true` + "\n"
//configuration += `}` + "\n"

// Note that RootDevicePath MUST use \\ not \ as path separator

func Create(ID string, Configuration string) error {

	log.Debugln("hcsshim::Create")
	log.Debugln("ID:", ID)
	log.Debugln("Configuration:", Configuration)

	var (
		// To pass into syscall, we need uint16 pointers to the strings
		IDp            *uint16
		Configurationp *uint16

		// Result values from calling the procedure
		r1, r2 uintptr

		// The DLL and procedure in the DLL respectively
		dll  *syscall.DLL
		proc *syscall.Proc

		// Error tracking
		err error
	)

	// Load the DLL and get the CreateComputeSystem function
	dll, proc, err = loadAndFind(PROCCREATE)

	// Release once used if we managed to get a handle to it
	if dll != nil {
		defer dll.Release()
	}

	// Check for error from loadAndFind
	if err != nil {
		return err
	}

	// Convert ID to uint16 pointers for calling the procedure
	IDp, err = syscall.UTF16PtrFromString(ID)
	if err != nil {
		log.Debugln("Failed conversion of ID to pointer ", err)
		return err
	}

	// Convert Configuration to uint16 pointers for calling the procedure
	Configurationp, err = syscall.UTF16PtrFromString(Configuration)
	if err != nil {
		log.Debugln("Failed conversion of Configuration to pointer ", err)
		return err
	}

	// Call the procedure itself.
	r1, r2, err = proc.Call(uintptr(unsafe.Pointer(IDp)),
		uintptr(unsafe.Pointer(Configurationp)))

	// Check the result codes first
	if r1 != 0 || r2 != 0 {
		log.Debugln("r1 ", r1)
		return errors.New(PROCCREATE + " failed r1/r2 check")
	}

	// Check for error itself next
	if err != nil {
		if err.Error() != "The operation completed successfully." {
			return errors.New(PROCCREATE + " failed. " + err.Error())
		}
	}

	return nil
} // Create

// Pass in hcsshim.start or hcsshim.stop
func ChangeState(ID string, newState int) error {

	log.Debugln("hcsshim::ChangeState")
	log.Debugln("ID:", ID)
	log.Debugln("State:", newState)

	var (
		// To pass into syscall, we need uint16 pointers to the strings
		IDp *uint16

		// Result values from calling the procedure
		r1, r2 uintptr

		// The DLL and procedure in the DLL respectively
		dll  *syscall.DLL
		proc *syscall.Proc

		// The name of the procedure
		procname string

		// Error tracking
		err error
	)

	switch newState {
	case Start:
		procname = PROCSTART
	case Stop:
		procname = PROCTERMINATE
	default:
		{
			return errors.New("Invalid newState passed to ChangeState()")
		}
	}

	// Load the DLL and get the function
	dll, proc, err = loadAndFind(procname)

	// Release once used if we managed to get a handle to it
	if dll != nil {
		defer dll.Release()
	}

	// Check for error from loadAndFind
	if err != nil {
		return err
	}

	// Convert ID to uint16 pointers for calling the procedure
	IDp, err = syscall.UTF16PtrFromString(ID)
	if err != nil {
		log.Debugln("Failed conversion of ID to pointer ", err)
		return err
	}

	// Call the procedure itself.
	r1, r2, err = proc.Call(uintptr(unsafe.Pointer(IDp)))

	// Check the result codes first
	if r1 != 0 || r2 != 0 {
		log.Debugln("r1 ", r1)
		return errors.New(procname + " failed r1/r2 check")
	}

	// Check for error itself next
	if err != nil {
		if err.Error() != "The operation completed successfully." {
			return errors.New(procname + " failed. " + err.Error())
		}
	}

	return nil
} // ChangeState

// RunAndWait runs a command inside a Windows container and waits for
// it to complete. The exit code of the command/process is returned back.
// It is possible to interact by passing named pipes through the devices
// parameter (eg docker run -a stdin -a stdout -a stderr -i container command)
func RunAndWait(ID string, CommandLine string, StdDevices Devices) (ExitCode uint32, err error) {

	log.Debugln("hcsshim::RunAndWait")
	log.Debugln("ID:", ID)
	log.Debugln("CommandLine:", CommandLine)

	var (
		// To pass into syscall, we need uint16 pointers to the strings
		IDp          *uint16
		CommandLinep *uint16

		// Result values from calling the procedure
		r1, r2 uintptr

		// The DLL and procedure in the DLL respectively
		dll  *syscall.DLL
		proc *syscall.Proc
	)

	// Load the DLL and get the ExecuteInComputeSystem function
	dll, proc, err = loadAndFind(PROCRUNANDWAIT)

	// Release once used if we managed to get a handle to it
	if dll != nil {
		defer dll.Release()
	}

	// Check for error from loadAndFind
	if err != nil {
		return 0, err
	}

	// Convert ID to uint16 pointer for calling the procedure
	IDp, err = syscall.UTF16PtrFromString(ID)
	if err != nil {
		log.Debugln("Failed conversion of ID to pointer ", err)
		return 0, err
	}

	// Convert CommandLine to uint16 pointer for calling the procedure
	CommandLinep, err = syscall.UTF16PtrFromString(CommandLine)
	if err != nil {
		log.Debugln("Failed conversion of CommandLine to pointer ", err)
		return 0, err
	}

	// Need an instance of the redirection devices for internal use when calling the procedure
	internalDevices := new(deviceInt)

	// Convert stdin, if supplied to uint16 pointer for calling the procedure
	if len(StdDevices.StdInPipe) != 0 {
		internalDevices.stdinpipe, err = syscall.UTF16PtrFromString(StdDevices.StdInPipe)
		if err != nil {
			log.Debugln("Failed conversion of StdInPipe to pointer ", err)
			return 0, err
		}
	}

	// Convert stdout, if supplied to uint16 pointer for calling the procedure
	if len(StdDevices.StdOutPipe) != 0 {
		internalDevices.stdoutpipe, err = syscall.UTF16PtrFromString(StdDevices.StdOutPipe)
		if err != nil {
			log.Debugln("Failed conversion of StdOutPipe to pointer ", err)
			return 0, err
		}
	}

	// Convert stderr, if supplied to uint16 pointer for calling the procedure
	if len(StdDevices.StdErrPipe) != 0 {
		internalDevices.stderrpipe, err = syscall.UTF16PtrFromString(StdDevices.StdErrPipe)
		if err != nil {
			log.Debugln("Failed conversion of StdErrPipe to pointer ", err)
			return 0, err
		}
	}

	// To get a POINTER to the exit code
	ec := new(uint32)

	// Call the procedure itself.
	r1, r2, err = proc.Call(
		uintptr(unsafe.Pointer(IDp)),
		uintptr(unsafe.Pointer(CommandLinep)),
		uintptr(unsafe.Pointer(internalDevices)),
		uintptr(unsafe.Pointer(ec)))

	// Check the result codes first
	if r1 != 0 || r2 != 0 {
		log.Debugln("r1 ", r1)
		return 0, errors.New(PROCRUNANDWAIT + " failed r1/r2 check")
	}

	// Check for error itself next
	if err != nil {
		if err.Error() != "The operation completed successfully." {
			return 0, errors.New(PROCRUNANDWAIT + " failed. " + err.Error())
		}
	}

	if ec != nil {
		log.Debugln("hcsshim::ExecuteInComputeSystem ExitCode ", *ec)
	}

	return *ec, nil
} // RunAndWait

func CreateProcessInComputeSystem(ID string, CommandLine string, StdDevices Devices) (PID uint32, err error) {

	log.Debugln("hcsshim::CreateProcessInComputeSystem")
	log.Debugln("ID:", ID)
	log.Debugln("CommandLine:", CommandLine)

	var (
		// To pass into syscall, we need uint16 pointers to the strings
		IDp          *uint16
		CommandLinep *uint16

		// Result values from calling the procedure
		r1, r2 uintptr

		// The DLL and procedure in the DLL respectively
		dll  *syscall.DLL
		proc *syscall.Proc
	)

	// Load the DLL and get the CreateProcessInComputeSystem function
	dll, proc, err = loadAndFind(PROCCREATEPROCESSINCOMPUTESYSTEM)

	// Release once used if we managed to get a handle to it
	if dll != nil {
		defer dll.Release()
	}

	// Check for error from loadAndFind
	if err != nil {
		return 0, err
	}

	// Convert ID to uint16 pointer for calling the procedure
	IDp, err = syscall.UTF16PtrFromString(ID)
	if err != nil {
		log.Debugln("Failed conversion of ID to pointer ", err)
		return 0, err
	}

	// Convert CommandLine to uint16 pointer for calling the procedure
	CommandLinep, err = syscall.UTF16PtrFromString(CommandLine)
	if err != nil {
		log.Debugln("Failed conversion of CommandLine to pointer ", err)
		return 0, err
	}

	// Need an instance of the redirection devices for internal use when calling the procedure
	internalDevices := new(deviceInt)

	// Convert stdin, if supplied to uint16 pointer for calling the procedure
	if len(StdDevices.StdInPipe) != 0 {
		internalDevices.stdinpipe, err = syscall.UTF16PtrFromString(StdDevices.StdInPipe)
		if err != nil {
			log.Debugln("Failed conversion of StdInPipe to pointer ", err)
			return 0, err
		}
	}

	// Convert stdout, if supplied to uint16 pointer for calling the procedure
	if len(StdDevices.StdOutPipe) != 0 {
		internalDevices.stdoutpipe, err = syscall.UTF16PtrFromString(StdDevices.StdOutPipe)
		if err != nil {
			log.Debugln("Failed conversion of StdOutPipe to pointer ", err)
			return 0, err
		}
	}

	// Convert stderr, if supplied to uint16 pointer for calling the procedure
	if len(StdDevices.StdErrPipe) != 0 {
		internalDevices.stderrpipe, err = syscall.UTF16PtrFromString(StdDevices.StdErrPipe)
		if err != nil {
			log.Debugln("Failed conversion of StdErrPipe to pointer ", err)
			return 0, err
		}
	}

	// To get a POINTER to the PID
	pid := new(uint32)

	// Call the procedure itself.
	r1, r2, err = proc.Call(
		uintptr(unsafe.Pointer(IDp)),
		uintptr(unsafe.Pointer(CommandLinep)),
		uintptr(unsafe.Pointer(internalDevices)),
		uintptr(unsafe.Pointer(pid)))

	// Check the result codes first
	if r1 != 0 || r2 != 0 {
		log.Debugln("r1 ", r1)
		return 0, errors.New(PROCCREATEPROCESSINCOMPUTESYSTEM + " failed r1/r2 check")
	}

	// Check for error itself next
	if err != nil {
		if err.Error() != "The operation completed successfully." {
			return 0, errors.New(PROCCREATEPROCESSINCOMPUTESYSTEM + " failed. " + err.Error())
		}
	}

	if pid != nil {
		log.Debugln("hcsshim::CreateProcessInComputeSystem PID ", *pid)
	}

	return *pid, nil
} // CreateProcessInComputeSystem

func loadAndFind(Procedure string) (dll *syscall.DLL, proc *syscall.Proc, err error) {

	log.Debugln("hcsshim::loadAndFind ", Procedure)

	dll, err = syscall.LoadDLL(SHIMDLL)
	if err != nil {
		log.Debugln("Failed to load ", SHIMDLL, err)
		return nil, nil, err
	}

	proc, err = dll.FindProc(Procedure)
	if err != nil {
		log.Debugln("Failed to find " + Procedure + " in " + SHIMDLL)
		return nil, nil, err
	}

	return dll, proc, err
}
