package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	specs_go "github.com/opencontainers/runtime-spec/specs-go"
	cp "github.com/otiai10/copy"
)

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
