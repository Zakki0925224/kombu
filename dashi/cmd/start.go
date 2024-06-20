package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/Zakki0925224/kombu/dashi/internal"
	"github.com/Zakki0925224/kombu/dashi/util"
	"github.com/charmbracelet/log"
	"github.com/google/subcommands"
)

type Start struct {
	mountSource string
	mountDest   string
}

func (t *Start) Name() string     { return "start" }
func (t *Start) Synopsis() string { return "start container" }
func (t *Start) Usage() string    { return "start <container-id> |<commands>|: " + t.Synopsis() }
func (t *Start) SetFlags(f *flag.FlagSet) {
	f.StringVar(&t.mountSource, "mount-source", "", "mount source path")
	f.StringVar(&t.mountDest, "mount-dest", "", "mount destination path")
}
func (t *Start) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()
	if len(args) == 0 {
		fmt.Printf("%s\n", t.Usage())
		return subcommands.ExitFailure
	}

	pSock, cSock, err := internal.NewPairSocket("syncsocket")
	if err != nil {
		log.Error("Error occured", "err", err)
		return subcommands.ExitFailure
	}
	defer pSock.Close()
	defer cSock.Close()

	r, err := internal.NewRuntime()
	if err != nil {
		log.Error("Error occured", "err", err)
		return subcommands.ExitFailure
	}
	c := r.FindContainer(args[0])
	if c == nil {
		log.Error("Container was not found", "cId", args[0])
		return subcommands.ExitFailure
	}

	if c.IsRunningContainer() {
		log.Error("Container is already running", "cId", args[0])
		return subcommands.ExitFailure
	}

	if !util.IsRunningRootUser() {
		c.ConvertSpecToRootless()
	}

	cmd := exec.Command(os.Args[0], "init")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = c.SpecSysProcAttr()
	cmd.ExtraFiles = []*os.File{cSock.F}
	if err := cmd.Start(); err != nil {
		log.Error("Error occured", "err", err)
		return subcommands.ExitFailure
	}

	opt := &internal.InitOption{
		Args:            args[1:],
		UserMountSource: t.mountSource,
		UserMountDest:   t.mountDest,
	}

	var mountList []string

	for {
		if pSock.IsClose() {
			break
		}

		bytes, err := pSock.Read()
		if err != nil {
			log.Error("Error occured", "err", err)
			pSock.Close()
		}
		req, err := internal.GetRequestFromBytes(bytes)
		if err != nil {
			log.Error("Error occured", "err", err)
			pSock.Close()
		}

		log.Debug("Received request from child", "req", req)

		switch req {
		case "close_con":
			pSock.Close()

		case "get_cid":
			if _, err := pSock.Write([]byte(c.Id)); err != nil {
				log.Error("Failed to send cId", "err", err)
				pSock.Close()
			}

		case "get_init_opt":
			bytes, err := json.Marshal(opt)
			if err != nil {
				log.Error("Failed to marshal init option", "err", err)
				pSock.Close()
			}
			if _, err := pSock.Write(bytes); err != nil {
				log.Error("Failed to send init option", "err", err)
				pSock.Close()
			}

		case "send_mount_list":
			bytes, err := pSock.Read()
			if err != nil {
				log.Error("Failed to receive mount list", "err", err)
				pSock.Close()
			}
			if err := json.Unmarshal(bytes, &mountList); err != nil {
				log.Error("Failed to unmarshal mount list", "err", err)
				pSock.Close()
			}

		case "unmount":
			if util.IsRunningRootUser() {
				c.Unmount(mountList)
			}
			if _, err := pSock.Write([]byte("ok")); err != nil {
				log.Error("Failed to send unmount response", "err", err)
				pSock.Close()
			}
		}
	}

	return subcommands.ExitSuccess
}
