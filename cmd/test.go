package cmd

import (
	"context"
	"flag"
	"fmt"

	"github.com/google/subcommands"
)

type Test struct {}

func (*Test) Name() string { return "test" }
func (*Test) Synopsis() string { return "test" }
func (*Test) Usage() string { return "test" }
func (*Test) SetFlags(f *flag.FlagSet) {
    // abc
}

func (t *Test) Execute(_ context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
    fmt.Println("Executed test command")
    return subcommands.ExitSuccess
}
