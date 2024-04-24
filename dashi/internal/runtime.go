package internal

import (
	"fmt"
	"io/fs"
	"os"
)

type Runtime struct {
	Containers []Container
}

func NewRuntime() (Runtime, error) {
	if err := os.MkdirAll(CONTAINERS_PATH, fs.ModePerm); err != nil {
		return Runtime{}, err
	}

	c, err := FindContainersFromDirectory()
	if err != nil {
		return Runtime{}, err
	}

	return Runtime{
		Containers: c,
	}, nil
}

func (r *Runtime) CreateContainer(ociRuntimeBundlePath string) (string, error) {
	c, err := NewContainer(ociRuntimeBundlePath)
	if err != nil {
		return "", err
	}

	r.Containers = append(r.Containers, c)
	return c.Id, nil
}

func (r *Runtime) DeleteContainer(cId string) error {
	cIdx := -1
	for i, c := range r.Containers {
		if c.Id == cId {
			cIdx = i
			break
		}
	}

	if cIdx != -1 {
		err := r.Containers[cIdx].DeleteContainerDirectory()
		if err != nil {
			return err
		}

		r.Containers = append(r.Containers[:cIdx], r.Containers[cIdx+1:]...)
	} else {
		return fmt.Errorf("Container was not found: %s", cId)
	}

	return nil
}

func (r *Runtime) FindContainer(cId string) *Container {
	cIdx := -1
	for i, c := range r.Containers {
		if c.Id == cId {
			cIdx = i
			break
		}
	}

	if cIdx == -1 {
		return nil
	}

	return &r.Containers[cIdx]
}
