package internal

import "fmt"

type Runtime struct {
	Containers []Container
}

func NewRuntime() Runtime {
	return Runtime{
		Containers: FindContainersFromDirectory(),
	}
}

func (r *Runtime) CreateContainer(ociRuntimeBundlePath string) string {
	c := NewContainer(ociRuntimeBundlePath)
	r.Containers = append(r.Containers, c)
	return c.Id
}

func (r *Runtime) DeleteContainer(cId string) {
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
			fmt.Printf("%v\n", err)
		}

		r.Containers = append(r.Containers[:cIdx], r.Containers[cIdx+1:]...)
		fmt.Printf("Deleted container: %s\n", cId)
	} else {
		fmt.Printf("Container was not found: %s\n", cId)
	}
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
