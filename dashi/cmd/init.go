package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/exec"

	"github.com/Zakki0925224/kombu/dashi/internal"
	"github.com/charmbracelet/log"
	"github.com/google/subcommands"
)

type Init struct{}

func (t *Init) Name() string             { return "init" }
func (t *Init) Synopsis() string         { return "container's init process" }
func (t *Init) Usage() string            { return "init: " + t.Synopsis() }
func (t *Init) SetFlags(f *flag.FlagSet) {}
func (t *Init) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	cSock := internal.GetSocketFromChild("syncsocket-c")
	defer func() {
		internal.RequestUnmount(cSock)
		bytes, _ := internal.RequestToBytes("close_con")
		cSock.Write(bytes)
		cSock.Close()
	}()

	// request and get init option
	var opt internal.InitOption
	bytes, _ := internal.RequestToBytes("get_init_opt")
	if _, err := cSock.Write(bytes); err != nil {
		log.Error("Failed to send request", "err", err)
		return subcommands.ExitFailure
	}
	bytes, err := cSock.Read()
	if err != nil {
		log.Error("Failed to receive response", "err", err)
		return subcommands.ExitFailure
	}
	if err := json.Unmarshal(bytes, &opt); err != nil {
		log.Error("Failed to unmarshal response", "err", err)
		return subcommands.ExitFailure
	}

	// call program
	cmd := exec.Command(opt.Args[0], opt.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Warn("Exit status was not 0", "err", err)
	}

	return subcommands.ExitSuccess
}
