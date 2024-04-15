package cmd

import (
	"context"
	"flag"
	"fmt"

	"github.com/Zakki0925224/kombu/internal"
	"github.com/google/subcommands"
)

type List struct{}

func (t *List) Name() string             { return "list" }
func (t *List) Synopsis() string         { return "show container list" }
func (t *List) Usage() string            { return "list: " + t.Synopsis() }
func (t *List) SetFlags(f *flag.FlagSet) {}
func (t *List) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	r := internal.NewRuntime()

	for i, c := range r.Containers {
		fmt.Printf("[%d]: %#v\n", i, c)
	}

	return subcommands.ExitSuccess
}
