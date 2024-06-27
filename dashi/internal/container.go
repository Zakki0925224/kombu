package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/charmbracelet/log"
	specs_go "github.com/opencontainers/runtime-spec/specs-go"
	cp "github.com/otiai10/copy"
	"golang.org/x/sys/unix"
)

type InitOption struct {
	Args            []string `json:args`
	UserMountSource string   `json:user_mount_source`
	UserMountDest   string   `json:user_mount_dest`
}

type Container struct {
	Id    string
	Spec  specs_go.Spec
	State specs_go.State
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
		Id:    containerId,
		Spec:  spec,
		State: state,
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
			Id:    info.Name(),
			Spec:  spec,
			State: state,
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

func (c *Container) Init(opt *InitOption) error {
	if err := c.SetSpecUid(); err != nil {
		return fmt.Errorf("Failed to set uid: %s", err)
	}

	if err := c.SetSpecGid(); err != nil {
		return fmt.Errorf("Failed to set gid: %s", err)
	}

	if err := c.SetSpecChroot(); err != nil {
		return fmt.Errorf("Failed to set chroot: %s", err)
	}

	if err := c.SetSpecChdir(); err != nil {
		return fmt.Errorf("Failed to set chdir: %s", err)
	}

	if err := c.SetSpecEnv(); err != nil {
		return fmt.Errorf("Failed to set env: %s", err)
	}

	if err := c.SetSpecHostname(); err != nil {
		return fmt.Errorf("Failed to set hostname: %s", err)
	}

	if err := c.SetSpecRlimits(); err != nil {
		return fmt.Errorf("Failed to set rlimits: %s", err)
	}

	if err := c.SetSpecCapabilities(); err != nil {
		return fmt.Errorf("Failed to set capabilities: %s", err)
	}

	return nil
}

func (c *Container) IsRunningContainer() bool {
	state := c.State
	return state.Status == specs_go.StateRunning && state.Pid != 1
}

// reference: https://github.com/opencontainers/runc/blob/main/libcontainer/specconv/example.go
func (c *Container) ConvertSpecToRootless() {
	var namespaces []specs_go.LinuxNamespace

	// remove networkns
	for _, ns := range c.Spec.Linux.Namespaces {
		switch ns.Type {
		case specs_go.NetworkNamespace, specs_go.UserNamespace:
			// do nothing
		default:
			namespaces = append(namespaces, ns)
		}
	}

	// add userns
	namespaces = append(namespaces, specs_go.LinuxNamespace{
		Type: specs_go.UserNamespace,
	})
	c.Spec.Linux.Namespaces = namespaces

	// add uid/gid mappings
	c.Spec.Linux.UIDMappings = []specs_go.LinuxIDMapping{{
		HostID:      uint32(os.Geteuid()),
		ContainerID: 0,
		Size:        1,
	}}
	c.Spec.Linux.GIDMappings = []specs_go.LinuxIDMapping{{
		HostID:      uint32(os.Getegid()),
		ContainerID: 0,
		Size:        1,
	}}

	// fix mounts
	var mounts []specs_go.Mount
	for _, mount := range c.Spec.Mounts {
		if filepath.Clean(mount.Destination) == "/sys" {
			mounts = append(mounts, specs_go.Mount{
				Source:      "/sys",
				Destination: "/sys",
				Type:        "none",
				Options:     []string{"rbind", "nosuid", "noexec", "nodev", "ro"},
			})
			continue
		}

		// remove all gid/uid mappings
		var options []string
		for _, option := range mount.Options {
			if !strings.HasPrefix(option, "gid=") && !strings.HasPrefix(option, "uid=") {
				options = append(options, option)
			}
		}

		mount.Options = options
		mounts = append(mounts, mount)
	}
	c.Spec.Mounts = mounts

	// remove cgroup settings
	c.Spec.Linux.Resources = nil
}

func (c *Container) SpecSysProcAttr() *syscall.SysProcAttr {
	flags := uintptr(syscall.CLONE_NEWNS)
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
		case specs_go.UserNamespace:
			flags |= syscall.CLONE_NEWUSER
		case specs_go.CgroupNamespace:
			flags |= syscall.CLONE_NEWCGROUP
		}
	}

	var uidMappings []syscall.SysProcIDMap
	var gidMappings []syscall.SysProcIDMap

	for _, mapping := range c.Spec.Linux.UIDMappings {
		uidMappings = append(uidMappings, syscall.SysProcIDMap{
			ContainerID: int(mapping.ContainerID),
			HostID:      int(mapping.HostID),
			Size:        int(mapping.Size),
		})
	}

	for _, mapping := range c.Spec.Linux.GIDMappings {
		gidMappings = append(gidMappings, syscall.SysProcIDMap{
			ContainerID: int(mapping.ContainerID),
			HostID:      int(mapping.HostID),
			Size:        int(mapping.Size),
		})
	}

	return &syscall.SysProcAttr{
		Cloneflags:  flags,
		UidMappings: uidMappings,
		GidMappings: gidMappings,
	}
}

func (c *Container) SetSpecRlimits() error {
	for _, res := range c.Spec.Process.Rlimits {
		if err := syscall.Setrlimit(rlimitMap[res.Type], &syscall.Rlimit{
			Cur: res.Soft,
			Max: res.Hard,
		}); err != nil {
			return err
		}
	}

	log.Debug("Set rlimits")
	return nil
}

func (c *Container) SetSpecChroot() error {
	rootFsPath := RootfsPath(c.Id, c.Spec.Root.Path)
	if err := syscall.Chroot(rootFsPath); err != nil {
		return err
	}

	log.Debug("Set root", "path", rootFsPath)
	return nil
}

func (c *Container) SetSpecChdir() error {
	cwd := c.Spec.Process.Cwd
	if err := syscall.Chdir(cwd); err != nil {
		return err
	}

	log.Debug("Set cwd", "cwd", cwd)
	return nil
}

func (c *Container) SetSpecHostname() error {
	hostname := c.Spec.Hostname
	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		return err
	}

	log.Debug("Set hostname", "hostname", hostname)
	return nil
}

func (c *Container) SetSpecMounts(opt *InitOption) ([]string, error) {
	rootFsPath := RootfsPath(c.Id, c.Spec.Root.Path)
	mounts := c.Spec.Mounts
	mountList := make([]string, 0)
	userMountSource := opt.UserMountSource
	userMountDest := opt.UserMountDest
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
			if f, ok := mountFlagMap[o]; ok {
				flags |= f
			}
		}

		// cgroup
		if source == mType && strings.Contains(source, "cgroup") {
			cVer := c.GetCgroupVersion()

			if cVer == 0 {
				log.Warn("Unsupported to mount cgroup")
				continue
			} else if cVer == 1 {
				source = "cgroup"
				mType = "tmpfs"
			} else {
				source = "cgroup2"
				mType = "cgroup2"
			}
		}

		if err := os.MkdirAll(dest, os.ModePerm); err != nil {
			log.Warn("Failed to create directory", "dest", dest, "err", err)
			//return subcommands.ExitFailure
		}

		if err := unix.Mount(source, dest, mType, flags, ""); err != nil {
			log.Warn("Failed to mount", "source", source, "err", err)
			//return subcommands.ExitFailure
		} else {
			log.Debug("Mounted", "source", source, "dest", dest)
			mountList = append(mountList, dest)
		}
	}

	return mountList, nil
}

func (c *Container) SetSpecCapabilities() error {
	caps := c.Spec.Process.Capabilities
	capHeader := unix.CapUserHeader{
		Version: unix.LINUX_CAPABILITY_VERSION_1,
		Pid:     int32(os.Getpid()),
	}

	eCaps := uint32(0)
	for _, e := range caps.Effective {
		if c, ok := capFlagMap[e]; ok && c < 32 {
			eCaps |= uint32(1 << c)
		}
	}

	pCaps := uint32(0)
	for _, e := range caps.Permitted {
		if c, ok := capFlagMap[e]; ok && c < 32 {
			pCaps |= uint32(1 << c)
		}
	}

	iCaps := uint32(0)
	for _, e := range caps.Inheritable {
		if c, ok := capFlagMap[e]; ok && c < 32 {
			iCaps |= uint32(1 << c)
		}
	}

	capData := unix.CapUserData{
		Effective:   eCaps,
		Permitted:   pCaps,
		Inheritable: iCaps,
	}

	if err := unix.Capset(&capHeader, &capData); err != nil {
		return err
	}

	log.Debug("Set capabilities", "caps", capData)
	return nil
}

func (c *Container) SetSpecUid() error {
	uid := int(c.Spec.Process.User.UID)
	if err := syscall.Setuid(uid); err != nil {
		return err
	}

	log.Debug("Set UID", "uid", uid)
	return nil
}

func (c *Container) SetSpecGid() error {
	gid := int(c.Spec.Process.User.GID)
	if err := syscall.Setgid(gid); err != nil {
		return err
	}

	log.Debug("Set GID", "gid", gid)
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

		log.Debug("Set env", "env", envKV)
	}

	return nil
}

func (c *Container) Unmount(mountList []string) {
	// sort nested paths
	sort.Slice(mountList, func(i, j int) bool {
		return len(strings.Split(mountList[i], "/")) > len(strings.Split(mountList[j], "/"))
	})

	for _, dest := range mountList {
		if err := unix.Unmount(dest, 0); err != nil {
			log.Warn("Failed to unmount", "dest", dest, "err", err)
		} else {
			log.Debug("Unmounted", "dest", dest)
		}
	}
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

// return 0: unsupported, 1: version1, 2: version2
func (c *Container) GetCgroupVersion() int {
	if _, err := os.Stat("/sys/fs/cgroup/cgroup.controllers"); err == nil {
		return 2
	} else if _, err := os.Stat("/sys/fs/cgroup"); err == nil {
		return 1
	} else {
		return 0
	}
}
