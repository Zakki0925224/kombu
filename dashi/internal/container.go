package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	specs_go "github.com/opencontainers/runtime-spec/specs-go"
	cp "github.com/otiai10/copy"
	"golang.org/x/sys/unix"
)

type StartOption struct {
	Args            []string
	UserMountSource string
	UserMountDest   string
}

type Container struct {
	Id        string
	Spec      specs_go.Spec
	State     specs_go.State
	MountList []string
}

func NewContainer(containerId string, ociRuntimeBundlePath string) (Container, error) {
	containerPath := ContainerPath(containerId)

	if _, err := os.Stat(containerPath); err == nil {
		return Container{}, fmt.Errorf("Container id already exists: %s", containerId)
	}

	// create container directory
	if err := os.MkdirAll(containerPath, os.ModePerm); err != nil {
		return Container{}, err
	}

	// copy config.json
	if err := cp.Copy(ociRuntimeBundlePath+"/"+BUNDLE_CONFIG_FILE_NAME, ConfigFilePath(containerId)); err != nil {
		os.RemoveAll(containerPath)
		return Container{}, err
	}

	// read config file
	bytes, err := os.ReadFile(ConfigFilePath(containerId))
	if err != nil {
		os.RemoveAll(containerPath)
		return Container{}, err
	}

	var spec specs_go.Spec
	if err := json.Unmarshal(bytes, &spec); err != nil {
		os.RemoveAll(containerPath)
		return Container{}, err
	}

	// copy rootfs
	if err := cp.Copy(ociRuntimeBundlePath+"/"+spec.Root.Path, RootfsPath(containerId, spec.Root.Path)); err != nil {
		os.RemoveAll(containerPath)
		return Container{}, err
	}

	// create state
	state := specs_go.State{
		Version:     spec.Version,
		ID:          containerId,
		Status:      specs_go.StateCreated,
		Pid:         -1,
		Bundle:      ociRuntimeBundlePath,
		Annotations: make(map[string]string),
	}

	// save state to file
	stateJson, err := json.Marshal(state)
	if err != nil {
		return Container{}, fmt.Errorf("Failed to convert state to json: %s", err)
	}

	file, err := os.Create(StateFilePath(containerId))
	if err != nil {
		return Container{}, fmt.Errorf("Failed to save state file: %s", err)
	}
	if _, err := file.Write(stateJson); err != nil {
		return Container{}, fmt.Errorf("Failed to save state file: %s", err)
	}

	if err := file.Close(); err != nil {
		return Container{}, fmt.Errorf("Failed to close state file: %s", err)
	}

	return Container{
		Id:        containerId,
		Spec:      spec,
		State:     state,
		MountList: make([]string, 0),
	}, nil
}

func FindContainersFromDirectory() ([]Container, error) {
	cs := make([]Container, 0)

	if _, err := os.Stat(CONTAINERS_PATH); os.IsNotExist(err) {
		return cs, err
	}

	var spec specs_go.Spec
	var state specs_go.State
	err := filepath.Walk(CONTAINERS_PATH, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return filepath.SkipDir
		}

		if path == CONTAINERS_PATH {
			return nil
		}

		if !info.IsDir() {
			return filepath.SkipDir
		}

		// read config file
		bytes, err := os.ReadFile(ConfigFilePath(info.Name()))
		if err != nil {
			return err
		}

		if err := json.Unmarshal(bytes, &spec); err != nil {
			return err
		}

		// read state file
		bytes, err = os.ReadFile(StateFilePath(info.Name()))
		if err != nil {
			return err
		}

		if err := json.Unmarshal(bytes, &state); err != nil {
			return err
		}

		cs = append(cs, Container{
			Id:        info.Name(),
			Spec:      spec,
			State:     state,
			MountList: make([]string, 0),
		})

		return filepath.SkipDir
	})

	if err != nil {
		return cs, err
	}

	return cs, nil
}

func (c *Container) DeleteContainerDirectory() error {
	return os.RemoveAll(CONTAINERS_PATH + "/" + c.Id)
}

func (c *Container) Save() error {
	// save state to file
	stateJson, err := json.Marshal(c.State)
	if err != nil {
		return fmt.Errorf("Failed to convert state to json: %s", err)
	}

	file, err := os.Create(StateFilePath(c.Id))
	if err != nil {
		return fmt.Errorf("Failed to create state file: %s", err)
	}
	if _, err := file.Write(stateJson); err != nil {
		return fmt.Errorf("Failed to write state file: %s", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("Failed to close state file: %s", err)
	}

	return nil
}

func (c *Container) Kill() error {
	if c.State.Status != specs_go.StateRunning && c.State.Pid == -1 {
		return fmt.Errorf("Container is not running")
	}

	// kill container process
	p, err := os.FindProcess(c.State.Pid)
	if err != nil {
		return err
	}
	if err := p.Kill(); err != nil {
		return err
	}

	c.State.Status = specs_go.StateStopped
	c.State.Pid = -1
	if err := c.Save(); err != nil {
		return err
	}

	return nil
}

func (c *Container) Start(opt *StartOption) error {
	if err := c.SetSpecHostname(); err != nil {
		return fmt.Errorf("Failed to set hostname: %s", err)
	}

	if err := c.SetSpecMounts(opt.UserMountSource, opt.UserMountDest); err != nil {
		return fmt.Errorf("Failed to set mounts: %s", err)
	}

	if err := c.SetSpecChroot(); err != nil {
		return fmt.Errorf("Failed to set chroot: %s", err)
	}

	if err := c.SetSpecChdir(); err != nil {
		return fmt.Errorf("Failed to set chdir: %s", err)
	}

	if err := c.SetSpecUid(); err != nil {
		return fmt.Errorf("Failed to set uid: %s", err)
	}

	if err := c.SetSpecGid(); err != nil {
		return fmt.Errorf("Failed to set gid: %s", err)
	}

	// if err := c.SetSpecCapabilities(); err != nil {
	// 	return fmt.Errorf("Failed to set capabilities: %s", err)
	// }

	if err := c.SetSpecEnv(); err != nil {
		return fmt.Errorf("Failed to set env: %s", err)
	}

	runArgs := c.Spec.Process.Args

	if opt != nil && len(opt.Args) > 0 {
		runArgs = opt.Args
	}

	fmt.Printf("Start container..., args: %v\n", runArgs)

	cmd := exec.Command(runArgs[0], runArgs[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to run command: %s\n", err)
	}

	if err := c.Unmount(); err != nil {
		return fmt.Errorf("Failed to unmount: %s", err)
	}

	return nil
}

func (c *Container) IsRunningContainer() bool {
	state := c.State
	return state.Status == specs_go.StateRunning && state.Pid != 1
}

func (c *Container) SpecNamespaceFlags() uintptr {
	flags := uintptr(0)
	for _, ns := range c.Spec.Linux.Namespaces {
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

	return flags
}

func (c *Container) SetSpecChroot() error {
	rootFsPath := RootfsPath(c.Id, c.Spec.Root.Path)
	if err := syscall.Chroot(rootFsPath); err != nil {
		return err
	}

	fmt.Printf("Set root: %s\n", rootFsPath)
	return nil
}

func (c *Container) SetSpecChdir() error {
	cwd := c.Spec.Process.Cwd
	if err := syscall.Chdir(cwd); err != nil {
		return err
	}

	fmt.Printf("Set cwd: %s\n", cwd)
	return nil
}

func (c *Container) SetSpecHostname() error {
	hostname := c.Spec.Hostname
	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		return err
	}

	fmt.Printf("Set hostname: %s\n", hostname)
	return nil
}

func (c *Container) SetSpecMounts(userMountSource string, userMountDest string) error {
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

	rootFsPath := RootfsPath(c.Id, c.Spec.Root.Path)

	mounts := c.Spec.Mounts
	if userMountSource != "" && userMountDest != "" {
		mounts = append(mounts, specs_go.Mount{
			Destination: userMountDest,
			Type:        "bind",
			Source:      userMountSource,
			Options:     []string{"rbind", "rw"},
		})
	}

	for _, m := range mounts {
		source := m.Source
		dest := rootFsPath + m.Destination
		mType := m.Type
		flags := uintptr(0)

		for _, o := range m.Options {
			if f, ok := mFlags[o]; ok {
				flags |= f
			}
			// else {
			// 	fmt.Printf("Undefined mount option: %s\n", o)
			// 	return subcommands.ExitFailure
			// }
		}

		if err := unix.Mount(source, dest, mType, flags, ""); err != nil {
			fmt.Printf("Failed to mount %s: %s\n", source, err)
			//return subcommands.ExitFailure
		} else {
			fmt.Printf("Mount %s to %s\n", source, dest)
			c.MountList = append(c.MountList, m.Destination)
		}
	}

	return nil
}

func (c *Container) SetSpecCapabilities() error {
	cFlags := map[string]uintptr{
		"CAP_AUDIT_CONTROL":      unix.CAP_AUDIT_CONTROL,
		"CAP_AUDIT_READ":         unix.CAP_AUDIT_READ,
		"CAP_AUDIT_WRITE":        unix.CAP_AUDIT_WRITE,
		"CAP_BLOCK_SUSPEND":      unix.CAP_BLOCK_SUSPEND,
		"CAP_BPF":                unix.CAP_BPF,
		"CAP_CHECKPOINT_RESTORE": unix.CAP_CHECKPOINT_RESTORE,
		"CAP_CHOWN":              unix.CAP_CHOWN,
		"CAP_DAC_OVERRIDE":       unix.CAP_DAC_OVERRIDE,
		"CAP_DAC_READ_SEARCH":    unix.CAP_DAC_READ_SEARCH,
		"CAP_FOWNER":             unix.CAP_FOWNER,
		"CAP_FSETID":             unix.CAP_FSETID,
		"CAP_IPC_LOCK":           unix.CAP_IPC_LOCK,
		"CAP_IPC_OWNER":          unix.CAP_IPC_OWNER,
		"CAP_KILL":               unix.CAP_KILL,
		"CAP_LEASE":              unix.CAP_LEASE,
		"CAP_LINUX_IMMUTABLE":    unix.CAP_LINUX_IMMUTABLE,
		"CAP_MAC_ADMIN":          unix.CAP_MAC_ADMIN,
		"CAP_MAC_OVERRIDE":       unix.CAP_MAC_OVERRIDE,
		"CAP_MKNOD":              unix.CAP_MKNOD,
		"CAP_NET_ADMIN":          unix.CAP_NET_ADMIN,
		"CAP_NET_BIND_SERVICE":   unix.CAP_NET_BIND_SERVICE,
		"CAP_NET_BROADCAST":      unix.CAP_NET_BROADCAST,
		"CAP_NET_RAW":            unix.CAP_NET_RAW,
		"CAP_PERFMON":            unix.CAP_PERFMON,
		"CAP_SETGID":             unix.CAP_SETGID,
		"CAP_SETFCAP":            unix.CAP_SETFCAP,
		"CAP_SETPCAP":            unix.CAP_SETPCAP,
		"CAP_SETUID":             unix.CAP_SETUID,
		"CAP_SYS_ADMIN":          unix.CAP_SYS_ADMIN,
		"CAP_SYS_BOOT":           unix.CAP_SYS_BOOT,
		"CAP_SYS_CHROOT":         unix.CAP_SYS_CHROOT,
		"CAP_SYS_MODULE":         unix.CAP_SYS_MODULE,
		"CAP_SYS_NICE":           unix.CAP_SYS_NICE,
		"CAP_SYS_PACCT":          unix.CAP_SYS_PACCT,
		"CAP_SYS_PTRACE":         unix.CAP_SYS_PTRACE,
		"CAP_SYS_RAWIO":          unix.CAP_SYS_RAWIO,
		"CAP_SYS_RESOURCE":       unix.CAP_SYS_RESOURCE,
		"CAP_SYS_TIME":           unix.CAP_SYS_TIME,
		"CAP_SYS_TTY_CONFIG":     unix.CAP_SYS_TTY_CONFIG,
		"CAP_SYSLOG":             unix.CAP_SYSLOG,
		"CAP_WAKE_ALARM":         unix.CAP_WAKE_ALARM,
	}

	caps := c.Spec.Process.Capabilities
	capData := unix.CapUserData{}
	capHeader := unix.CapUserHeader{
		Version: unix.LINUX_CAPABILITY_VERSION_3,
		Pid:     int32(c.State.Pid),
	}

	eCaps := uint32(0)
	for _, e := range caps.Effective {
		if c, ok := cFlags[e]; ok {
			eCaps |= uint32(c)
		}
	}
	capData.Effective = eCaps

	pCaps := uint32(0)
	for _, e := range caps.Permitted {
		if c, ok := cFlags[e]; ok {
			pCaps |= uint32(c)
		}
	}
	capData.Permitted = pCaps

	iCaps := uint32(0)
	for _, e := range caps.Inheritable {
		if c, ok := cFlags[e]; ok {
			iCaps |= uint32(c)
		}
	}
	capData.Inheritable = iCaps

	if err := unix.Capset(&capHeader, &capData); err != nil {
		return err
	}

	fmt.Printf("Set capabilities: %#v\n", capData)
	return nil
}

func (c *Container) SetSpecUid() error {
	uid := int(c.Spec.Process.User.UID)
	if err := syscall.Setuid(uid); err != nil {
		return err
	}

	fmt.Printf("Set UID: %d\n", uid)
	return nil
}

func (c *Container) SetSpecGid() error {
	gid := int(c.Spec.Process.User.GID)
	if err := syscall.Setgid(gid); err != nil {
		return err
	}

	fmt.Printf("Set GID: %d\n", gid)
	return nil
}

func (c *Container) SetSpecEnv() error {
	for _, envKV := range c.Spec.Process.Env {
		kv := strings.Split(envKV, "=")
		key := kv[0]
		value := kv[1]
		if err := os.Setenv(key, value); err != nil {
			return err
		}

		fmt.Printf("Set env: %s\n", envKV)
	}

	return nil
}

func (c *Container) Unmount() error {
	for _, dest := range c.MountList {
		if err := unix.Unmount(dest, 0); err != nil {
			fmt.Printf("Failed to unmount %s: %s\n", dest, err)
		}
	}

	c.MountList = make([]string, 0)
	return nil
}

func (c *Container) SetStateRunning() error {
	c.State.Status = specs_go.StateRunning
	c.State.Pid = os.Getpid()

	return c.Save()
}

func (c *Container) SetStateStopped() error {
	c.State.Status = specs_go.StateStopped
	c.State.Pid = -1

	return c.Save()
}
