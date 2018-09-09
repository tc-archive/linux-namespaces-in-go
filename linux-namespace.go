package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	// 'reexec'  provides a convenient way for an executable to “re-exec” itself.
	// It is required to circumvent a limitation in how Go handles process forking.
	// *exec.Cmd.Run() has no mechanism to allow new namespace properties to be
	// altered before the new proccess is creates. reexec allows an 'initialisation'
	// function to be specified that is invoked before the new process is created.
	//
	// Installed via dep: 'dep ensure -add github.com/docker/docker/pkg/reexec'
	//
	reexec "github.com/docker/docker/pkg/reexec"
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
	// NB: USe rexec instead of exec to allow safe pre-initialisation of namespaces.
	rootfsPath := "/tmp/ns-process/rootfs"
	cmd := reexec.Command("nsInitialisation", rootfsPath)
	// cmd := exec.Command("/bin/sh")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = []string{"PS1=-[ns-process]- # "}

	// Create SysProcID mappings to provide root (uid=0,gid=0) status inside
	// the container.
	// ```cat /etc/passwd | awk -F: '{printf "%s:%s:%s\n",$1,$3,$4}'````
	uidMappings, gidMapMappings := createSysProcIDMappings(0, 0)

	// allows us to run code after the namespace creation but before the process starts. This is where reexec comes in.

	// SysProcAttr allows attributes to be set on commands.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:
		// When requesting a new User namespace alongside other namespaces, the
		// User namespace will be created first. User namespaces can be created without
		// root permissions, which means sudo can now be dropped when the new process
		// is created.
		//
		// A new User namespace - but by default the process will have no uid, gid, or groups.
		// -[ns-process]- # id || echo $USER || whoami
		syscall.CLONE_NEWUSER |
			// A new Mount namespace - but by default the process will use the host's mounts and rootfs.
			// -[ns-process]- # ls / || car /proc
			syscall.CLONE_NEWNS |
			// A new UTS namespace - but by default the process will will use the host's hostname and domainname.
			// -[ns-process]- # hostname && domainname
			syscall.CLONE_NEWUTS |
			// A new IPC namespace - but by default the process will have no messages queues, shared memory segments, or, semaphore arrays..
			// -[ns-process]- # ipcs
			syscall.CLONE_NEWIPC |
			// A new UTS namespace - but by default the process will only have the loopback interface.
			// -[ns-process]- # ip link show || route || ping 10.10.10.1
			syscall.CLONE_NEWNET |
			// A new PID namespace - but by default a new /proc filesystem has not be mounted.
			// -[ns-process]- # ls /proc
			syscall.CLONE_NEWPID,
		UidMappings: uidMappings,
		GidMappings: gidMapMappings,
	}

	// cmd.Run() invokes the clone() syscall. The specified Cloneflags are used
	// to modify the behaviour of the clone() operation.
	// This controls which namespaces the process to be executed in.
	//
	// Instead of running the command in one step, we Start() it and Wait() for
	// any initialisation to complete before we continue.
	//
	// if err := cmd.Run(); err != nil {
	// 	fmt.Printf("Error running the /bin/sh command - %s\n", err)
	// 	os.Exit(1)
	// }

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting the reexec.Command - %s\n", err)
		os.Exit(1)
	}

	// run netsetgo using default args
	// note that netsetgo must be owned by root with the setuid bit set
	pid := fmt.Sprintf("%d", cmd.Process.Pid)
	netsetgoPath := "./assets/netsetgo"
	netsetgoCmd := exec.Command(netsetgoPath, "-pid", pid)
	if err := netsetgoCmd.Run(); err != nil {
		fmt.Printf("Error running netsetgo - %s\n", err)
		os.Exit(1)
	}

	if err := cmd.Wait(); err != nil {
		fmt.Printf("Error waiting for the reexec.Command - %s\n", err)
		os.Exit(1)
	}
}

// Reexec Execution ***********************************************************

// An initialisation function for rexec.
//
func nsInitialisation() {
	fmt.Printf("\n>> initialising namespace <<\n\n")

	newrootPath := os.Args[1]

	// Mount a new proc filesystem
	if err := mountProc(newrootPath); err != nil {
		fmt.Printf("Error mounting /proc - %s\n", err)
		os.Exit(1)
	}

	// Pivot root on new mountspace.
	if err := pivotRoot(newrootPath); err != nil {
		fmt.Printf("Error running pivot_root - %s\n", err)
		os.Exit(1)
	}

	// Set the hostname
	if err := syscall.Sethostname([]byte("ns-process")); err != nil {
		fmt.Printf("Error setting hostname - %s\n", err)
		os.Exit(1)
	}

	// Wait for new network bridge device to come up.
	if err := waitForNetwork(); err != nil {
		fmt.Printf("Error waiting for network - %s\n", err)
		os.Exit(1)
	}

	nsRun()
}

// Register the the default initialisation function.
//
func init() {
	reexec.Register("nsInitialisation", nsInitialisation)
	if reexec.Init() {
		// Prevents infinite loop initialisation.
		os.Exit(0)
	}
}

// Execute the rexec function.
//
func nsRun() {
	cmd := exec.Command("/bin/sh")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = []string{"PS1=-[ns-process]- # "}

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running the /bin/sh command - %s\n", err)
		os.Exit(1)
	}
}
