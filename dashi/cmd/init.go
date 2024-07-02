package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/exec"
	"strconv"
	"time"

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
		bytes, _ := internal.RequestToBytes("close_con")
		cSock.Write(bytes)
		cSock.Close()
	}()

	ticker := time.NewTicker(10 * time.Second) // 10 ticker
	defer ticker.Stop()

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

	if err := cmd.Start(); err != nil {
		log.Error("Failed to start program", "err", err)
		return subcommands.ExitFailure
	}

	// socket connection check
	targetProgramPid := -1
	go func() {
		pingReqBytes, _ := internal.RequestToBytes("ping")
		sendTargetProgramPidReqBytes, _ := internal.RequestToBytes("send_target_program_pid")
		for range ticker.C {
			if targetProgramPid != -1 {
				cSock.Write(sendTargetProgramPidReqBytes)
				cSock.Write([]byte(strconv.Itoa(targetProgramPid)))
				targetProgramPid = -1
			} else {
				cSock.Write(pingReqBytes)
			}
		}
	}()

	// get pid
	targetProgramPid = cmd.Process.Pid

	if err := cmd.Wait(); err != nil {
		log.Warn("Exit status was not 0", "err", err)
	}

	return subcommands.ExitSuccess
}
