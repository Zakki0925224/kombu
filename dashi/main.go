package main

import (
	"context"
	"flag"
	"os"

	"github.com/Zakki0925224/kombu/dashi/cmd"
	"github.com/charmbracelet/log"
	"github.com/google/subcommands"
)

func main() {
	log.SetLevel(log.DebugLevel)

	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(new(cmd.PreInit), "")
	subcommands.Register(new(cmd.Init), "")
	subcommands.Register(new(cmd.Start), "")
	subcommands.Register(new(cmd.Create), "")
	subcommands.Register(new(cmd.Delete), "")
	subcommands.Register(new(cmd.List), "")
	subcommands.Register(new(cmd.Download), "")
	subcommands.Register(new(cmd.Kill), "")
	flag.Parse()

	ctx := context.Background()
	status := subcommands.Execute(ctx)
	os.Exit(int(status))
}
