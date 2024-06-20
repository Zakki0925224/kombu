package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/Zakki0925224/kombu/dashi/internal"
	"github.com/Zakki0925224/kombu/dashi/util"
	"github.com/charmbracelet/log"
	"github.com/google/subcommands"
)

type Init struct{}

func (t *Init) Name() string             { return "init" }
func (t *Init) Synopsis() string         { return "initialize container" }
func (t *Init) Usage() string            { return "init: " + t.Synopsis() }
func (t *Init) SetFlags(f *flag.FlagSet) {}
func (t *Init) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	cSock := internal.GetSocketFromChild("syncsocket-c")
	defer func() {
		t.RequestUnmount(cSock)
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

	if c.IsRunningContainer() {
		log.Error("Container is already running", "cId", string(cId))
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

	// if err := internal.SetKeepCaps(); err != nil {
	// 	log.Error("Failed to set keep capabilities", "err", err)
	// 	return subcommands.ExitFailure
	// }

	// TODO: capability is not inherited
	if err := c.Start(opt.Args); err != nil {
		log.Error("Failed to start container", "err", err)
		return subcommands.ExitFailure
	}

	// if err := internal.ClearKeepCaps(); err != nil {
	// 	log.Error("Failed to clear keep capabilities", "err", err)
	// 	return subcommands.ExitFailure
	// }

	return subcommands.ExitSuccess
}

func (t *Init) RequestUnmount(s *internal.SyncSocket) error {
	bytes, _ := internal.RequestToBytes("unmount")
	if _, err := s.Write(bytes); err != nil {
		return err
	}

	// receive response
	bytes, err := s.Read()
	if err != nil {
		return err
	}

	if string(bytes) != "ok" {
		return fmt.Errorf("Failed to unmount")
	}

	return nil
}
