package cmd

import (
	"context"
	"flag"
	"fmt"

	"github.com/Zakki0925224/kombu/internal"
	"github.com/google/subcommands"
)

type Delete struct{}

func (*Delete) Name() string             { return "delete" }
func (*Delete) Synopsis() string         { return "delete container" }
func (*Delete) Usage() string            { return "delete <container-id>: delete container" }
func (*Delete) SetFlags(f *flag.FlagSet) {}
func (t *Delete) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()

	if len(args) != 1 {
		fmt.Printf("%s\n", t.Usage())
		return subcommands.ExitFailure
	}

	r := internal.NewRuntime()
	r.DeleteContainer(args[0])
	return subcommands.ExitSuccess
}
