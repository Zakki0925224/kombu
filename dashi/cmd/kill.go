package cmd

import (
	"context"
	"flag"
	"fmt"

	"github.com/Zakki0925224/kombu/dashi/internal"
	"github.com/google/subcommands"
)

type Kill struct{}

func (t *Kill) Name() string             { return "kill" }
func (t *Kill) Synopsis() string         { return "kill running container" }
func (t *Kill) Usage() string            { return "kill <container-id>: " + t.Synopsis() }
func (t *Kill) SetFlags(f *flag.FlagSet) {}
func (t *Kill) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()

	if len(args) != 1 {
		fmt.Printf("%s\n", t.Usage())
		return subcommands.ExitFailure
	}

	r, err := internal.NewRuntime()
	if err != nil {
		fmt.Printf("Error occured: %s\n", err)
		return subcommands.ExitFailure
	}

	if err := r.KillRunningContainer(args[0]); err != nil {
		fmt.Printf("Error occured: %s\n", err)
		return subcommands.ExitFailure
	}

	fmt.Printf("Successed to kill container process\n")
	return subcommands.ExitSuccess
}
