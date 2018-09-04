package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// Namespaces provide a way to limit what a process can see, to make it appear
// as though it is the only process running on a host.
//
// The Namespaces API
//
// The namespaces(7) man page defines 3 system calls that make up the API:
//
// clone(2) - Creates a new process.
// setns(2) - Allows the calling process to join an existing namespace.
// unshare(2) - Moves the calling process to a new namespace.
//
// Namespace | Constant        | Isolates
// ----------+-----------------+-----------------------------------------------
// Cgroup    | CLONE_NEWCGROUP | Isolate cgroup root directory.
// IPC       | CLONE_NEWIPC    | Isolate IPC resources, POSIX message queues.
// Network   | CLONE_NEWNET    | Isolate network devices, stacks, ports, etc.
// Mount     | CLONE_NEWNS     | Isolate filesystem mount points.
// PID       | CLONE_NEWPID    | Process PID number space.
// User      | CLONE_NEWUSER   | Isolate UID/GID number spaces.
// UTS       | CLONE_NEWUTS    | Isolate hostname and NIS domainname.
//
//
// Root privileges are required to create most namespaces (except the 'User'
// namespace).
//
func main() {

	fmt.Println("Running linux-namespace...")

	// Create a new shell process with io streams and custom prompt.
	cmd := exec.Command("/bin/sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = []string{"PS1=-[ns-process]- # "}
	// SysProcAttr allows attributes to be set on commands.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// The UTS namespace provides isolation of the 'hostname' and 'domainname'
		// system identifiers.
		//
		// Changing the hostname inside the new shell does not affect the hostname
		// of the calling process.
		Cloneflags: syscall.CLONE_NEWUTS,
	}

	// cmd.Run() invokes the clone() syscall. The specified Cloneflags are used
	// to modify the behaviour of the clone() operation.
	// This controls which namespaces the process to be executed in.
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running the /bin/sh command - %s\n", err)
		os.Exit(1)
	}
}
