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

	PROCCREATECOMPUTESYSTEM             = "CreateComputeSystem"
	PROCSTARTCOMPUTESYSTEM              = "StartComputeSystem"
	PROCCREATEPROCESSINCOMPUTESYSTEM    = "CreateProcessInComputeSystem"
	PROCWAITFORPROCESSINCOMPUTESYSTEM   = "WaitForProcessInComputeSystem"
	PROCSHUTDOWNCOMPUTESYSTEM           = "ShutdownComputeSystem"
	PROCTERMINATEPROCESSINCOMPUTESYSTEM = "TerminateProcessInComputeSystem"
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

func CreateComputeSystem(ID string, Configuration string) error {

	log.Debugln("hcsshim::CreateComputeSystem")
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
	dll, proc, err = loadAndFind(PROCCREATECOMPUTESYSTEM)

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
		// Check for error itself next
		if err != nil {
			if err.Error() != "The operation completed successfully." {
				return errors.New(PROCCREATECOMPUTESYSTEM + " failed. " + err.Error())
			}
		}
		log.Debugln("r1 ", r1)
		return errors.New(PROCCREATECOMPUTESYSTEM + " failed r1/r2 check")
	}

	return nil
} // CreateComputeSystem

func StartComputeSystem(ID string) error {

	log.Debugln("hcsshim::StartComputeSystem")
	log.Debugln("ID:", ID)

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

	// Load the DLL and get the function
	dll, proc, err = loadAndFind(PROCSTARTCOMPUTESYSTEM)

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
		// Check for error itself next
		if err != nil {
			if err.Error() != "The operation completed successfully." {
				return errors.New(PROCSTARTCOMPUTESYSTEM + " failed. " + err.Error())
			}
		}
		return errors.New(procname + " failed r1/r2 check")
	}

	return nil
} // StartComputeSystem

func CreateProcessInComputeSystem(ID string, CommandLine string, StdDevices Devices, EmulateTTY uint32) (PID uint32, err error) {

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
		uintptr(EmulateTTY),
		uintptr(unsafe.Pointer(pid)))

	// Check the result codes first
	if r1 != 0 || r2 != 0 {
		log.Debugln("r1 ", r1)
		// Check for error itself next
		if err != nil {
			if err.Error() != "The operation completed successfully." {
				return 0, errors.New(PROCCREATEPROCESSINCOMPUTESYSTEM + " failed. " + err.Error() + ". Command=" + CommandLine)
			}
		}

		return 0, errors.New(PROCCREATEPROCESSINCOMPUTESYSTEM + " could not run " + CommandLine)
	}

	if pid != nil {
		log.Debugln("hcsshim::CreateProcessInComputeSystem PID ", *pid)
	}

	return *pid, nil
} // CreateProcessInComputeSystem

func WaitForProcessInComputeSystem(ID string, ProcessId uint32) (ExitCode uint32, err error) {

	log.Debugln("hcsshim::WaitForProcessInComputeSystem")
	log.Debugln("ID:", ID)
	log.Debugln("ProcessID:", ProcessId)

	var (
		// To pass into syscall, we need uint16 pointers to the strings
		IDp *uint16

		// Result values from calling the procedure
		r1, r2 uintptr

		// The DLL and procedure in the DLL respectively
		dll  *syscall.DLL
		proc *syscall.Proc

		// Infinite
		Timeout uint32 = 0xFFFFFFFF // (-1)
	)

	// Load the DLL and get the CreateProcessInComputeSystem function
	dll, proc, err = loadAndFind(PROCWAITFORPROCESSINCOMPUTESYSTEM)

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

	// To get a POINTER to the ExitCode
	ec := new(uint32)

	// Call the procedure itself.
	r1, r2, err = proc.Call(
		uintptr(unsafe.Pointer(IDp)),
		uintptr(ProcessId),
		uintptr(Timeout),
		uintptr(unsafe.Pointer(ec)))

	// Check the result codes first
	if r1 != 0 || r2 != 0 {
		log.Debugln("r1 ", r1)
		// Check for error itself next
		if err != nil {
			if err.Error() != "The operation completed successfully." {
				return 0, errors.New(PROCWAITFORPROCESSINCOMPUTESYSTEM + " failed. " + err.Error())
			}
		}

		return 0, errors.New(PROCWAITFORPROCESSINCOMPUTESYSTEM + " failed r1/r2 check")
	}

	if ec != nil {
		log.Debugln("hcsshim::WaitForProcessInComputeSystem ExitCode ", *ec)
	}

	return *ec, nil
} // WaitForProcessInComputeSystem

func TerminateProcessInComputeSystem(ID string, ProcessId uint32) (err error) {

	log.Debugln("hcsshim::TerminateProcessInComputeSystem")
	log.Debugln("ID:", ID)
	log.Debugln("ProcessID:", ProcessId)

	var (
		// To pass into syscall, we need uint16 pointers to the strings
		IDp *uint16

		// Result values from calling the procedure
		r1, r2 uintptr

		// The DLL and procedure in the DLL respectively
		dll  *syscall.DLL
		proc *syscall.Proc
	)

	// Load the DLL and get the TerminateProcessInComputeSystem function
	dll, proc, err = loadAndFind(PROCTERMINATEPROCESSINCOMPUTESYSTEM)

	// Release once used if we managed to get a handle to it
	if dll != nil {
		defer dll.Release()
	}

	// Check for error from loadAndFind
	if err != nil {
		return err
	}

	// Convert ID to uint16 pointer for calling the procedure
	IDp, err = syscall.UTF16PtrFromString(ID)
	if err != nil {
		log.Debugln("Failed conversion of ID to pointer ", err)
		return err
	}

	// Call the procedure itself.
	r1, r2, err = proc.Call(
		uintptr(unsafe.Pointer(IDp)),
		uintptr(ProcessId))

	// Check the result codes first
	if r1 != 0 || r2 != 0 {
		log.Debugln("r1 ", r1)

		// Check for error itself next
		if err != nil {
			if err.Error() != "The operation completed successfully." {
				return errors.New(PROCTERMINATEPROCESSINCOMPUTESYSTEM + " failed. " + err.Error())
			}
		}

		return errors.New(PROCTERMINATEPROCESSINCOMPUTESYSTEM + " failed r1/r2 check")

	}

	return nil
} // TerminateProcessInComputeSystem

func ShutdownComputeSystem(ID string) error {

	log.Debugln("hcsshim::ShutdownComputeSystem")
	log.Debugln("ID:", ID)

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

	// Load the DLL and get the function
	dll, proc, err = loadAndFind(PROCSHUTDOWNCOMPUTESYSTEM)

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

	var timeout uint32
	timeout = 0xffffffff

	// Call the procedure itself.
	r1, r2, err = proc.Call(uintptr(unsafe.Pointer(IDp)), uintptr(timeout))

	// Check the result codes first
	if r1 != 0 || r2 != 0 {
		log.Debugln("r1 ", r1)
		// Check for error itself next
		if err != nil {
			if err.Error() != "The operation completed successfully." {
				return errors.New(PROCSHUTDOWNCOMPUTESYSTEM + " failed. " + err.Error())
			}
		}
		return errors.New(procname + " failed r1/r2 check")
	}

	return nil
} // ShutdownComputeSystem

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
