package main

import (
	"context"
	"flag"
	"os"

	"github.com/Zakki0925224/kombu/cmd"
	"github.com/google/subcommands"
)

func main() {
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(new(cmd.Run), "")
	subcommands.Register(new(cmd.Create), "")
	subcommands.Register(new(cmd.Delete), "")
	subcommands.Register(new(cmd.List), "")
	flag.Parse()

	ctx := context.Background()
	status := subcommands.Execute(ctx)
	os.Exit(int(status))
}
