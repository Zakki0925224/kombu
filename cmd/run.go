package cmd

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os/exec"

	"github.com/google/subcommands"
)

type Run struct{}

func (*Run) Name() string             { return "run" }
func (*Run) Synopsis() string         { return "run" }
func (*Run) Usage() string            { return "run" }
func (*Run) SetFlags(f *flag.FlagSet) {}
func (t *Run) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()

	if len(args) == 0 {
		fmt.Printf("%s\n", t.Usage())
		return subcommands.ExitFailure
	}

	cmd := exec.Command(args[0], args[1:]...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	err = cmd.Start()
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	fmt.Printf("Running at pid: %d\n", cmd.Process.Pid)

	scan := bufio.NewScanner(stdout)
	scanErr := bufio.NewScanner(stderr)
	go print(scan)
	go print(scanErr)

	err = cmd.Wait()
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

func print(r *bufio.Scanner) {
	for r.Scan() {
		fmt.Println(r.Text())
	}
}
