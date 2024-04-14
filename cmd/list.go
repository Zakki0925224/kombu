package cmd

import (
	"context"
	"flag"
	"fmt"

	"github.com/Zakki0925224/kombu/internal"
	"github.com/google/subcommands"
)

type List struct{}

func (*List) Name() string             { return "list" }
func (*List) Synopsis() string         { return "show container list" }
func (*List) Usage() string            { return "list: show container list" }
func (*List) SetFlags(f *flag.FlagSet) {}
func (t *List) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	r := internal.NewRuntime()

	for i, c := range r.Containers {
		fmt.Printf("[%d]: %s\n", i, c.Id)
	}

	return subcommands.ExitSuccess
}
