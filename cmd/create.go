package cmd

import (
	"context"
	"flag"

	"github.com/Zakki0925224/kombu/internal"
	"github.com/google/subcommands"
)

type Create struct{}

func (*Create) Name() string             { return "create" }
func (*Create) Synopsis() string         { return "create new container" }
func (*Create) Usage() string            { return "create: Create new container" }
func (*Create) SetFlags(f *flag.FlagSet) {}
func (t *Create) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	r := internal.NewRuntime()
	r.CreateContainer()
	return subcommands.ExitSuccess
}
