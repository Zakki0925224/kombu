package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

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

	if !util.IsRunningRootUser() {
		c.ConvertSpecToRootless()
	}

	cmd := exec.Command(os.Args[0], "preinit")
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
		Args:            c.Spec.Process.Args,
		UserMountSource: t.mountSource,
		UserMountDest:   t.mountDest,
	}

	if args[1:] != nil && len(args[1:]) > 0 {
		opt.Args = args[1:]
	}

	var mountList []string

	checkPing := false
	pingReceived := make(chan bool)
	pingTimeoutDur := 15 * time.Second
	targetProgramPid := -1

	go func() {
		for {
			if !checkPing {
				continue
			}

			select {
			case <-pingReceived:
				continue
			case <-time.After(pingTimeoutDur):
				log.Error("syncsocket connection timed out, container process zombied")
				pSock.Close()

				if targetProgramPid != -1 {
					// TODO: pid is invalid for host
					// kill target program
					// if err := syscall.Kill(targetProgramPid, syscall.SIGKILL); err != nil {
					// 	log.Error("Failed to kill target program", "pid", targetProgramPid, "err", err)
					// }

					// log.Info("Target program killed")
				}
				return
			}
		}
	}()

	for {
		if pSock.IsClose() {
			if util.IsRunningRootUser() {
				c.Unmount(mountList)
			}
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

		//log.Debug("Received request from child", "req", req)

		switch req {
		case "ping":
			checkPing = true
			pingReceived <- true

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

		case "send_target_program_pid":
			bytes, err := pSock.Read()
			if err != nil {
				log.Error("Failed to receive target program pid", "err", err)
				continue
			}
			pid, err := strconv.Atoi(string(bytes))
			if err != nil {
				log.Error("Failed to parse target program pid", "err", err)
				continue
			}
			targetProgramPid = pid
		}
	}

	log.Info("Exited container")
	return subcommands.ExitSuccess
}
