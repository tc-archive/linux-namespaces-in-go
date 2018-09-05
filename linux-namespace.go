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
		Cloneflags:
		// When requesting a new User namespace alongside other namespaces, the
		// User namespace will be created first. User namespaces can be created without
		// root permissions, which means sudo can now be dropped when the new process
		// is created.
		//
		// A new User namespace - but by default the process will have no uid, gid, or groups.
		// -[ns-process]- # id
		syscall.CLONE_NEWUSER |
			// A new Mount namespace - but by default the process will use the host's mounts and rootfs.
			// -[ns-process]- # ls /
			syscall.CLONE_NEWNS |
			// A new UTS namespace - but by default the process will only have the loopback interface.
			// -[ns-process]- # ip link show
			syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWIPC |
			// A new UTS namespace - but by default the process will only have the loopback interface.
			// -[ns-process]- # ip link show
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWPID,
	}

	// We’ve requested a new PID namespace (CLONE_NEWPID) but haven't mounted a new /proc filesystem

	// We’ve requested a new User namespace (CLONE_NEWUSER) but have failed to provide a UID/GID mapping

	// cmd.Run() invokes the clone() syscall. The specified Cloneflags are used
	// to modify the behaviour of the clone() operation.
	// This controls which namespaces the process to be executed in.
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running the /bin/sh command - %s\n", err)
		os.Exit(1)
	}
}
