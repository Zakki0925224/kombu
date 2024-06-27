package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"syscall"

	"github.com/Zakki0925224/kombu/dashi/internal"
	"github.com/Zakki0925224/kombu/dashi/util"
	"github.com/charmbracelet/log"
	"github.com/google/subcommands"
)

// preinit is container's host
// call init prosess by syscall.Exec
type PreInit struct{}

func (t *PreInit) Name() string             { return "preinit" }
func (t *PreInit) Synopsis() string         { return "pre initialize container" }
func (t *PreInit) Usage() string            { return "preinit: " + t.Synopsis() }
func (t *PreInit) SetFlags(f *flag.FlagSet) {}
func (t *PreInit) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	cSock := internal.GetSocketFromChild("syncsocket-c")
	defer func() {
		bytes, _ := internal.RequestToBytes("close_con")
		cSock.Write(bytes)
		cSock.Close()
	}()

	// request and get container id
	bytes, _ := internal.RequestToBytes("get_cid")
	if _, err := cSock.Write(bytes); err != nil {
		log.Error("Failed to send request", "err", err)
		return subcommands.ExitFailure
	}
	cId, err := cSock.Read()
	if err != nil {
		log.Error("Failed to receive response", "err", err)
		return subcommands.ExitFailure
	}

	r, err := internal.NewRuntime()
	if err != nil {
		log.Error("Error occured", "err", err)
		return subcommands.ExitFailure
	}
	c := r.FindContainer(string(cId))
	if c == nil {
		log.Error("Container was not found", "cId", string(cId))
		return subcommands.ExitFailure
	}

	// request and get init option
	var opt internal.InitOption
	bytes, _ = internal.RequestToBytes("get_init_opt")
	if _, err := cSock.Write(bytes); err != nil {
		log.Error("Failed to send request", "err", err)
		return subcommands.ExitFailure
	}
	bytes, err = cSock.Read()
	if err != nil {
		log.Error("Failed to receive response", "err", err)
		return subcommands.ExitFailure
	}
	if err := json.Unmarshal(bytes, &opt); err != nil {
		log.Error("Failed to unmarshal response", "err", err)
		return subcommands.ExitFailure
	}

	if !util.IsRunningRootUser() {
		c.ConvertSpecToRootless()
	}

	// send mount list
	mountList, err := c.SetSpecMounts(&opt)
	if err != nil {
		log.Error("Failed to set mounts", "err", err)
		return subcommands.ExitFailure
	}
	bytes, _ = internal.RequestToBytes("send_mount_list")
	if _, err := cSock.Write(bytes); err != nil {
		log.Error("Failed to send request", "err", err)
		return subcommands.ExitFailure
	}
	listBytes, err := json.Marshal(mountList)
	if err != nil {
		log.Error("Failed to marshal mount list", "err", err)
		return subcommands.ExitFailure
	}
	if _, err := cSock.Write(listBytes); err != nil {
		log.Error("Failed to send mount list", "err", err)
		return subcommands.ExitFailure
	}

	if err := c.Init(&opt); err != nil {
		log.Error("Failed to initialize container", "err", err)
		return subcommands.ExitFailure
	}

	// TODO: capability is not inherited
	log.Info("Start container...", "args", opt.Args)

	if err := syscall.Exec("/proc/self/exe", []string{"/proc/self/exe", "init"}, os.Environ()); err != nil {
		log.Error("Failed to exec init", "err", err)
		return subcommands.ExitFailure
	}
	// unreachable here
	return subcommands.ExitSuccess
}
