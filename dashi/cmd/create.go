package cmd

import (
	"context"
	"flag"
	"fmt"

	"github.com/Zakki0925224/kombu/dashi/internal"
	"github.com/charmbracelet/log"
	"github.com/google/subcommands"
)

type Create struct{}

func (t *Create) Name() string     { return "create" }
func (t *Create) Synopsis() string { return "create new container" }
func (t *Create) Usage() string {
	return "create <container-id> <OCI runtime bundle path>: " + t.Synopsis()
}
func (t *Create) SetFlags(f *flag.FlagSet) {}
func (t *Create) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()

	if len(args) != 2 {
		fmt.Printf("%s\n", t.Usage())
		return subcommands.ExitFailure
	}

	r, err := internal.NewRuntime()
	if err != nil {
		log.Error("Error occured", "err", err)
		return subcommands.ExitFailure
	}

	cId, err := r.CreateContainer(args[0], args[1])
	if err != nil {
		log.Error("Error occured", "err", err)
		return subcommands.ExitFailure
	}

	log.Info("Created new container", "cId", cId)
	return subcommands.ExitSuccess
}
