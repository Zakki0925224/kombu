package cmd

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/Zakki0925224/kombu/internal"
	"github.com/google/subcommands"
)

type Run struct{}

func (*Run) Name() string             { return "run" }
func (*Run) Synopsis() string         { return "run command in the container" }
func (*Run) Usage() string            { return "run <container-id> <command...>" }
func (*Run) SetFlags(f *flag.FlagSet) {}
func (t *Run) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()

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

	// create child process
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	// create new namespace
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start process: %s\n", err)
		return subcommands.ExitFailure
	}
	pid := cmd.Process.Pid
	fmt.Printf("Running at pid: %d\n", pid)

	go handleStdin(stdin)
	go handleStdout(stdout)
	go handleStdout(stderr)

	// wait child process
	if err := cmd.Wait(); err != nil {
		fmt.Printf("Failed to wait process: %s\n", err)
		return subcommands.ExitFailure
	}

	fmt.Printf("Exited container\n")
	return subcommands.ExitSuccess
}

func handleStdin(f io.WriteCloser) {
	defer f.Close()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		t := scanner.Text()
		_, err := io.WriteString(f, t+"\n")

		if err != nil {
			fmt.Printf("Error writing to stdin: %s\n", err)
			break
		}
	}
}

func handleStdout(f io.ReadCloser) {
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fmt.Printf("%s\n", scanner.Text())
	}
}
