package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/Zakki0925224/kombu/internal"
	"github.com/google/subcommands"
	specs_go "github.com/opencontainers/runtime-spec/specs-go"
)

const SELF_PROC_PATH string = "/proc/self/exe"

type Run struct {
	child bool
}

func (t *Run) Name() string     { return "run" }
func (t *Run) Synopsis() string { return "run command in the container" }
func (t *Run) Usage() string    { return "run <container-id> <command...>: " + t.Synopsis() }
func (t *Run) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&t.child, "child", false, "run as child process")
}
func (t *Run) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()

	if !t.child {
		return t.execParent(args)
	}

	return t.execChild(args)
}

func (t *Run) execParent(args []string) subcommands.ExitStatus {
	// execute self binary instead of fork
	cmd := exec.Command(SELF_PROC_PATH, append([]string{"run", "--child"}, args[0:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	r, err := internal.NewRuntime()
	if err != nil {
		fmt.Printf("Error occured: %s\n", err)
		return subcommands.ExitFailure
	}

	c := r.FindContainer(args[0])
	if c == nil {
		fmt.Printf("Container was not found: %s\n", args[0])
		return subcommands.ExitFailure
	}

	spec := c.Spec

	// create new namespace
	flags := 0
	for _, ns := range spec.Linux.Namespaces {
		switch ns.Type {
		case specs_go.PIDNamespace:
			flags |= syscall.CLONE_NEWPID
		case specs_go.NetworkNamespace:
			flags |= syscall.CLONE_NEWNET
		case specs_go.IPCNamespace:
			flags |= syscall.CLONE_NEWIPC
		case specs_go.UTSNamespace:
			flags |= syscall.CLONE_NEWUTS
		case specs_go.MountNamespace:
			flags |= syscall.CLONE_NEWNS
		}
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: uintptr(flags),
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to run container: %s\n", err)
		return subcommands.ExitFailure
	}

	fmt.Printf("Exited container\n")
	return subcommands.ExitSuccess
}

func (t *Run) execChild(args []string) subcommands.ExitStatus {
	if len(args) < 2 {
		fmt.Printf("%s\n", t.Usage())
		return subcommands.ExitFailure
	}

	cId := args[0]
	cmdArgs := args[1:]

	r, err := internal.NewRuntime()
	if err != nil {
		fmt.Printf("Error occured: %s\n", err)
		return subcommands.ExitFailure
	}

	c := r.FindContainer(cId)
	if c == nil {
		fmt.Printf("Container was not found: %s\n", cId)
		return subcommands.ExitFailure
	}

	spec := c.Spec

	// set rootfs
	syscall.Chroot(internal.RootfsPath(cId))
	syscall.Chdir(spec.Process.Cwd)
	// set hostname
	syscall.Sethostname([]byte(spec.Hostname))
	// mount proc
	//syscall.Mount("proc", "proc", "proc", 0, "")

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	return subcommands.ExitSuccess
}
