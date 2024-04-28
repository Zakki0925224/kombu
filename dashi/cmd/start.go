package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/Zakki0925224/kombu/dashi/internal"
	"github.com/google/subcommands"
)

const SELF_PROC_PATH string = "/proc/self/exe"

type Start struct {
	child bool
}

func (t *Start) Name() string     { return "start" }
func (t *Start) Synopsis() string { return "start container" }
func (t *Start) Usage() string    { return "start <container-id>: " + t.Synopsis() }
func (t *Start) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&t.child, "child", false, "start container as child process")
}
func (t *Start) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()

	if !t.child {
		return t.execParent(args)
	}

	return t.execChild(args)
}

func (t *Start) execParent(args []string) subcommands.ExitStatus {
	if len(args) != 1 {
		fmt.Printf("%s\n", t.Usage())
		return subcommands.ExitFailure
	}

	// execute self binary instead of fork
	cmd := exec.Command(SELF_PROC_PATH, append([]string{"start", "--child"}, args[0:]...)...)
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

	if c.IsRunningContainer() {
		fmt.Printf("Container is running: %s\n", args[0])
		return subcommands.ExitFailure
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: c.SpecNamespaceFlags(),
	}

	if err := c.SetStateRunning(); err != nil {
		fmt.Printf("Failed to set container state: %s\n", err)
		return subcommands.ExitFailure
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to start container: %s\n", err)
	}

	if err := c.SetStateStopped(); err != nil {
		fmt.Printf("Failed to set container state: %s\n", err)
		return subcommands.ExitFailure
	}

	fmt.Printf("Exited container\n")
	return subcommands.ExitSuccess
}

func (t *Start) execChild(args []string) subcommands.ExitStatus {
	cId := args[0]

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

	if err := c.SetSpecChroot(); err != nil {
		fmt.Printf("Failed to set chroot: %s\n", err)
		return subcommands.ExitFailure
	}

	if err := c.SetSpecChdir(); err != nil {
		fmt.Printf("Failed to set chdir: %s\n", err)
		return subcommands.ExitFailure
	}

	if err := c.SetSpecHostname(); err != nil {
		fmt.Printf("Failed to set hostname: %s\n", err)
		return subcommands.ExitFailure
	}

	if err := c.SetSpecMounts(); err != nil {
		fmt.Printf("Failed to mount: %s\n", err)
		return subcommands.ExitFailure
	}
	// TODO: set users

	if err := c.SetSpecUid(); err != nil {
		fmt.Printf("Failed to set uid: %s\n", err)
		return subcommands.ExitFailure
	}

	if err := c.SetSpecGid(); err != nil {
		fmt.Printf("Failed to set gid: %s\n", err)
		return subcommands.ExitFailure
	}

	if err := c.SetSpecCapabilities(); err != nil {
		fmt.Printf("Failed to set capabilities: %s\n", err)
		return subcommands.ExitFailure
	}

	if err := c.SetSpecEnv(); err != nil {
		fmt.Printf("Failed to set env: %s\n", err)
		return subcommands.ExitFailure
	}

	c.Start()
	if err := c.Unmount(); err != nil {
		fmt.Printf("Failed to unmount: %s\n", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
