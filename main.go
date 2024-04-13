package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/docker/docker/pkg/reexec"
	"github.com/vandit1604/dockerium/docker"
)

// init is called before main function. This automatically registers the commands inside the initialisation
func init() {
	// via `Register` we can register the functions that we will use inside the namespace that we are creating for a container
	reexec.Register("initialisation", initialisation)
	// via Init we check if the registered function was actually exec'd or not
	if reexec.Init() {
		os.Exit(0)
	}
}

// does all the Initialisation of the namespace
func initialisation() {
	log.Printf("\n>> ANYTHING THAT WE WANT TO DO INSIDE THE NAMESPACE <<\n")
	newRootPath := os.Args[1]
	memoryLimit := os.Args[2]
	cpuLimit := os.Args[3]

	os.MkdirAll(newRootPath, 0700)

	if err := limitCPUandMemory(newRootPath, memoryLimit, cpuLimit); err != nil {
		log.Printf("Error Limiting Cpu and Memory - %s\n", err)
		os.Exit(1)
	}

	if err := mountProc(newRootPath); err != nil {
		log.Printf("Error mounting /proc - %s\n", err)
		os.Exit(1)
	}
	defer syscall.Unmount(filepath.Join(newRootPath, "/proc"), 0)

	if err := pivotRoot(newRootPath); err != nil {
		log.Printf("Error running pivot_root - %s\n", err)
		os.Exit(1)
	}

	// set hostname
	if err := syscall.Sethostname([]byte("dockerium")); err != nil {
		log.Fatalf("Error while changing the hostname inside the container: %v", err)
	}

	nsRun()
}

// It runs the shell inside the namespace after the initialisation
func nsRun() {
	cmd := exec.Command("/bin/sh")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// setting the env for better distinction between the namespaces
	// setting `TERM=xterm` to use clear command in the container
	cmd.Env = []string{"PS1=-[dockerium]- # ", "TERM=xterm"}

	// running the command
	if err := cmd.Run(); err != nil {
		log.Fatalf(`Error running the command: %v
			- Did you extract the assets/alpine-minirootfs-3.19.1-x86_64.tar.gz`, err)
		os.Exit(cmd.ProcessState.ExitCode())
	}
}

func main() {
	token := docker.Authenticate("debian")
	rootfsPath := "/tmp/dockerium/rootfs"
	memorylimit := "524288000"
	cpulimit := "512"

	//  We’re now passing an argument, rootfsPath, to initialisation.
	cmd := reexec.Command("initialisation", rootfsPath, memorylimit, cpulimit)

	// pipe the stdin/out/err of os to cmd
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	/*
					FOR GENERAL INFORMATION: https://man7.org/linux/man-pages/man7/namespaces.7.html
					syscall.CLONE_NEWUTS - new uts namespace which gives us a namespaced hostname and domain name to the process
					syscall.CLONE_NEWPID - new process id for this process in the namespace it will be 1
					syscall.CLONE_NEWIPC - https://www.man7.org/linux/man-pages/man7/ipc_namespaces.7.html
					syscall.CLONE_NEWNET - https://man7.org/linux/man-pages/man7/network_namespaces.7.html
					syscall.CLONE_NEWUSER - https://man7.org/linux/man-pages/man7/user_namespaces.7.html

		CLONE_NEWNS: https://man7.org/linux/man-pages/man7/mount_namespaces.7.html
		This flag has the same effect as the clone(2) CLONE_NEWNS
		flag.  Unshare the mount namespace, so that the calling
		process has a private copy of its namespace which is not
		shared with any other process.  Specifying this flag
		automatically implies CLONE_FS as well.  Use of
		CLONE_NEWNS requires the CAP_SYS_ADMIN capability.  For
		further information, see mount_namespaces(7).

		We're creating a new usernamespace which enables us to run the program not as a root user too
		ALTHOUGH, we're not mapping the user id's to the new namespace, so the user will not be root in the new namespace
	*/
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS |
			syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWUSER, // these mappings will give the new user in user namespace a root identity
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		},
	}

	// starting the command
	if err := cmd.Start(); err != nil {
		log.Fatalf("Error while starting the command %v:", err)
		os.Exit(cmd.ProcessState.ExitCode())
	}

	// waiting for the command
	if err := cmd.Wait(); err != nil {
		log.Fatalf(`Error running the command: %v
			- Did you extract the assets/alpine-minirootfs-3.19.1-x86_64.tar.gz`, err)
		os.Exit(cmd.ProcessState.ExitCode())
	}
}
