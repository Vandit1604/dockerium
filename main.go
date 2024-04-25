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

var (
	rootfsPath  string = "/tmp/dockerium/rootfs"
	memorylimit string = "524288000"
	cpulimit    string = "512"
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

	if err := limitCPUandMemory(rootfsPath, memorylimit, cpulimit); err != nil {
		log.Printf("Error Limiting Cpu and Memory - %s\n", err)
		os.Exit(1)
	}

	if err := mountProc(rootfsPath); err != nil {
		log.Printf("Error mounting /proc - %s\n", err)
		os.Exit(1)
	}
	defer syscall.Unmount(filepath.Join(rootfsPath, "/proc"), 0)

	if err := pivotRoot(rootfsPath); err != nil {
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
	/*
		Authenticate - see the authentication section of the documentation.
		Fetch the manifest for the image you wish to download. See the section Pulling an Image Manifest for details.
		Parse the manifest to identify the layers to be downloaded. See the Image Manifest V 2, Schema 2 for details of the fields in the manifest.
		Fetch each layer listed in the manifest. See the section on Pulling a Layer for details.
		Unzip the layers on top of each other to re-create the filesystem. Remember that the layer list is ordered starting from the base image.
		Fetch the config data and store it ready for Step 8.
	*/
	os.MkdirAll(rootfsPath, 0700)

	// fetch image here
	image := "debian"

	token, err := docker.Authenticate(image)
	if err != nil {
		log.Fatalf("Error authenticating for the image: %v", err)
	}

	manifest, err := docker.FetchManifest(image, token)
	if err != nil {
		log.Fatalf("Error fetching manifest: %v", err)
	}

	err = docker.FetchLayers(image, token, *manifest)
	if err != nil {
		log.Fatalf("Error fetching layers: %v", err)
	}

	config, err := docker.FetchConfig(image, token, *manifest)
	if err != nil {
		log.Fatalf("Error fetching layers: %v", err)
	}

	log.Println(config)

	//  We’re now passing the arguments, rootfsPath, memorylimit, cpulimit, to initialisation.
	cmd := reexec.Command("initialisation")

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
