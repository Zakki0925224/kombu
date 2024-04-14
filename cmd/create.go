package cmd

import (
	"context"
	"flag"
	"fmt"

	"github.com/Zakki0925224/kombu/internal"
	"github.com/google/subcommands"
)

type Create struct{}

func (*Create) Name() string             { return "create" }
func (*Create) Synopsis() string         { return "create new container" }
func (*Create) Usage() string            { return "create <rootfs-dir-path>: Create new container" }
func (*Create) SetFlags(f *flag.FlagSet) {}
func (t *Create) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()

	if len(args) != 1 {
		fmt.Printf("%s\n", t.Usage())
		return subcommands.ExitFailure
	}

	r := internal.NewRuntime()
	cId := r.CreateContainer(args[0])

	fmt.Printf("Created new container: %s\n", cId)
	return subcommands.ExitSuccess
}
