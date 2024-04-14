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
)

const SELF_PROC_PATH string = "/proc/self/exe"

type Run struct {
	child bool
}

func (*Run) Name() string     { return "run" }
func (*Run) Synopsis() string { return "run command in the container" }
func (*Run) Usage() string    { return "run <container-id> <command...>" }
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

	// create new namespace
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
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

	r := internal.NewRuntime()
	c := r.FindContainer(cId)

	if c == nil {
		fmt.Printf("Container was not found: %s\n", cId)
		return subcommands.ExitFailure
	}

	// set rootfs
	syscall.Chroot(internal.RootfsPath(cId))
	syscall.Chdir("/")
	// set hostname
	syscall.Sethostname([]byte("container-" + cId))
	// mount proc
	//syscall.Mount("proc", "proc", "proc", 0, "")

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	return subcommands.ExitSuccess
}
