package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/Zakki0925224/kombu/dashi/internal"
	"github.com/google/subcommands"
	specs_go "github.com/opencontainers/runtime-spec/specs-go"
	"golang.org/x/sys/unix"
)

const SELF_PROC_PATH string = "/proc/self/exe"

type Start struct {
	child bool
}

func (t *Start) Name() string     { return "start" }
func (t *Start) Synopsis() string { return "start container" }
func (t *Start) Usage() string    { return "start <container-id>: " + t.Synopsis() }
func (t *Start) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&t.child, "child", false, "start container as child process")
}
func (t *Start) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	args := f.Args()

	if !t.child {
		return t.execParent(args)
	}

	return t.execChild(args)
}

func (t *Start) execParent(args []string) subcommands.ExitStatus {
	if len(args) != 1 {
		fmt.Printf("%s\n", t.Usage())
		return subcommands.ExitFailure
	}

	// execute self binary instead of fork
	cmd := exec.Command(SELF_PROC_PATH, append([]string{"start", "--child"}, args[0:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	r, err := internal.NewRuntime()
	if err != nil {
		fmt.Printf("Error occured: %s\n", err)
		return subcommands.ExitFailure
	}

	c := r.FindContainer(args[0])
	if c == nil {
		fmt.Printf("Container was not found: %s\n", args[0])
		return subcommands.ExitFailure
	}

	spec := c.Spec

	// create new namespace
	flags := 0
	for _, ns := range spec.Linux.Namespaces {
		switch ns.Type {
		case specs_go.PIDNamespace:
			flags |= syscall.CLONE_NEWPID
		case specs_go.NetworkNamespace:
			flags |= syscall.CLONE_NEWNET
		case specs_go.IPCNamespace:
			flags |= syscall.CLONE_NEWIPC
		case specs_go.UTSNamespace:
			flags |= syscall.CLONE_NEWUTS
		case specs_go.MountNamespace:
			flags |= syscall.CLONE_NEWNS
		}
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: uintptr(flags),
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to start container: %s\n", err)
		return subcommands.ExitFailure
	}

	fmt.Printf("Exited container\n")
	return subcommands.ExitSuccess
}

func (t *Start) execChild(args []string) subcommands.ExitStatus {
	cId := args[0]

	r, err := internal.NewRuntime()
	if err != nil {
		fmt.Printf("Error occured: %s\n", err)
		return subcommands.ExitFailure
	}

	c := r.FindContainer(cId)
	if c == nil {
		fmt.Printf("Container was not found: %s\n", cId)
		return subcommands.ExitFailure
	}

	spec := c.Spec

	rootFsPath := internal.RootfsPath(cId, spec.Root.Path)
	if err := syscall.Chroot(rootFsPath); err != nil {
		fmt.Printf("Failed to syscall.chroot: %s\n", err)
		return subcommands.ExitFailure
	}
	fmt.Printf("Set root: %s\n", rootFsPath)

	cwd := spec.Process.Cwd
	if err := syscall.Chdir(cwd); err != nil {
		fmt.Printf("Failed to syscall.chdir: %s\n", err)
		return subcommands.ExitFailure
	}
	fmt.Printf("Set cwd: %s\n", cwd)

	if err := syscall.Sethostname([]byte(spec.Hostname)); err != nil {
		fmt.Printf("Failed to syscall.sethostname: %s\n", err)
		return subcommands.ExitFailure
	}
	fmt.Printf("Set hostname: %s\n", spec.Hostname)

	mFlags := map[string]uintptr{
		"async":         unix.MS_ASYNC,
		"atime":         0,
		"bind":          unix.MS_BIND,
		"defaults":      0,
		"dev":           unix.MS_NODEV,
		"diratime":      0,
		"dirsync":       unix.MS_DIRSYNC,
		"exec":          0,
		"iversion":      unix.MS_I_VERSION,
		"lazytime":      unix.MS_LAZYTIME,
		"loud":          0,
		"noatime":       unix.MS_NOATIME,
		"nodev":         unix.MS_NODEV,
		"nodiratime":    unix.MS_NODIRATIME,
		"noexec":        unix.MS_NOEXEC,
		"noiversion":    0,
		"nolazytime":    unix.MS_RELATIME,
		"norelatime":    unix.MS_RELATIME,
		"nostrictatime": unix.MS_RELATIME,
		"nosuid":        unix.MS_NOSUID,
		"private":       unix.MS_PRIVATE,
		"rbind":         unix.MS_BIND | unix.MS_REC,
		"rdev":          unix.MS_NODEV | unix.MS_REC,
		"rdiratime":     0 | unix.MS_REC,
		"relatime":      unix.MS_RELATIME,
		"remount":       unix.MS_REMOUNT,
		"ro":            unix.MS_RDONLY,
		"rprivate":      unix.MS_PRIVATE | unix.MS_REC,
		"rshared":       unix.MS_SHARED | unix.MS_REC,
		"rslave":        unix.MS_SLAVE | unix.MS_REC,
		"runbindable":   unix.MS_UNBINDABLE | unix.MS_REC,
		"rw":            0,
		"shared":        unix.MS_SHARED,
		"silent":        0,
		"slave":         unix.MS_SLAVE,
		"strictatime":   unix.MS_STRICTATIME,
		"suid":          0,
		"sync":          unix.MS_SYNC,
		"unbindable":    unix.MS_UNBINDABLE,
	}

	for _, m := range spec.Mounts {
		source := m.Source
		dest := m.Destination
		mType := m.Type
		flag := uintptr(0)

		for _, o := range m.Options {
			if f, ok := mFlags[o]; ok {
				flag |= f
			}
			// else {
			// 	fmt.Printf("Undefined mount option: %s\n", o)
			// 	return subcommands.ExitFailure
			// }
		}

		if err := unix.Mount(source, dest, mType, flag, ""); err != nil {
			fmt.Printf("Failed to mount %s: %s\n", source, err)
			//return subcommands.ExitFailure
		}
		fmt.Printf("Mount %s to %s\n", source, dest)
	}

	// TODO: set capabilities
	// TODO: set users

	// set UID / GID
	uid := int(spec.Process.User.UID)
	if err := syscall.Setuid(uid); err != nil {
		fmt.Printf("Failed to syscall.setuid: %s\n", err)
		return subcommands.ExitFailure
	}
	fmt.Printf("Set UID: %d\n", uid)

	gid := int(spec.Process.User.GID)
	if err := syscall.Setuid(gid); err != nil {
		fmt.Printf("Failed to syscall.setgid: %s\n", err)
		return subcommands.ExitFailure
	}
	fmt.Printf("Set GID: %d\n", gid)

	// set env
	for _, envKV := range spec.Process.Env {
		kv := strings.Split(envKV, "=")
		key := kv[0]
		value := kv[1]
		os.Setenv(key, value)
		fmt.Printf("Set env: %s\n", envKV)
	}

	runArgs := spec.Process.Args
	fmt.Printf("Start container..., args: %v\n", runArgs)

	cmd := exec.Command(runArgs[0], runArgs[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	return subcommands.ExitSuccess
}
