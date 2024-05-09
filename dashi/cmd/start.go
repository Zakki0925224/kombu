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
	child       bool
	user        bool
	mountSource string
	mountDest   string
}

func (t *Start) Name() string     { return "start" }
func (t *Start) Synopsis() string { return "start container" }
func (t *Start) Usage() string    { return "start <container-id> |<commands>|: " + t.Synopsis() }
func (t *Start) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&t.child, "child", false, "start container as child process")
	f.BoolVar(&t.user, "user", false, "start container as rootless")
	f.StringVar(&t.mountSource, "mount-source", "", "mount source path")
	f.StringVar(&t.mountDest, "mount-dest", "", "mount destination path")
}
func (t *Start) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()

	if !t.child {
		return t.execParent(args)
	}

	return t.execChild(args)
}

func (t *Start) execParent(args []string) subcommands.ExitStatus {
	if len(args) == 0 {
		fmt.Printf("%s\n", t.Usage())
		return subcommands.ExitFailure
	}

	newArgs := []string{"start", "--child"}
	if t.mountSource != "" {
		newArgs = append(newArgs, "-mount-source="+t.mountSource)
	}
	if t.mountDest != "" {
		newArgs = append(newArgs, "-mount-dest="+t.mountDest)
	}
	if t.user {
		newArgs = append(newArgs, "-user")
	}
	newArgs = append(newArgs, args[0:]...)

	// execute self binary instead of fork
	cmd := exec.Command(SELF_PROC_PATH, newArgs...)
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

	if (t.mountSource != "" && t.mountDest == "") ||
		(t.mountSource == "" && t.mountDest != "") {
		fmt.Printf("Invalid flags\n")
		return subcommands.ExitFailure
	}

	startOption := &internal.StartOption{
		Args:            args[1:],
		UserMountSource: t.mountSource,
		UserMountDest:   t.mountDest,
		User:            t.user,
	}

	if err := c.Start(startOption); err != nil {
		fmt.Printf("Failed to start container: %s\n", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
