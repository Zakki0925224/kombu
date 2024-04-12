package internal

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type Container struct {
	Id string
}

func NewContainer() Container {
	uuidObj, _ := uuid.NewRandom()
	uuidStr := uuidObj.String()

	// create container directory
	os.MkdirAll(CONTAINERS_PATH+"/"+uuidStr, os.ModePerm)

	return Container{
		Id: uuidStr,
	}
}

func FindContainersFromDirectory() []Container {
	cs := make([]Container, 0)

	if _, err := os.Stat(CONTAINERS_PATH); os.IsNotExist(err) {
		return cs
	}

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

		cs = append(cs, Container{
			Id: info.Name(),
		})

		return nil
	})

	if err != nil {
		panic(err)
	}

	return cs
}

func (c *Container) DeleteContainerDirectory() error {
	return os.RemoveAll(CONTAINERS_PATH + "/" + c.Id)
}
