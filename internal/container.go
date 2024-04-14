package internal

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
	cp "github.com/otiai10/copy"
)

type Container struct {
	Id string
}

func NewContainer(rootfsDirPath string) Container {
	uuidObj, _ := uuid.NewRandom()
	uuidStr := uuidObj.String()

	// create container and rootfs directory
	p := CONTAINERS_PATH + "/" + uuidStr + "/" + ROOTFS_DIR_NAME
	os.MkdirAll(p, os.ModePerm)
	// copy rootfs
	cp.Copy(rootfsDirPath, p)

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

		return filepath.SkipDir
	})

	if err != nil {
		panic(err)
	}

	return cs
}

func (c *Container) DeleteContainerDirectory() error {
	return os.RemoveAll(CONTAINERS_PATH + "/" + c.Id)
}
