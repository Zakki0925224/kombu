package internal

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
	cp "github.com/otiai10/copy"
)

type Container struct {
	Id         string
	ConfigJson string
}

func NewContainer(ociRuntimeBundlePath string) Container {
	uuidObj, _ := uuid.NewRandom()
	uuidStr := uuidObj.String()

	// create container and rootfs directory
	os.MkdirAll(RootfsPath(uuidStr), os.ModePerm)
	// copy rootfs
	cp.Copy(ociRuntimeBundlePath+"/"+ROOTFS_DIR_NAME, RootfsPath(uuidStr))
	// copy config.json
	cp.Copy(ociRuntimeBundlePath+"/"+BUNDLE_CONFIG_FILE_NAME, ConfigFilePath(uuidStr))
	// read config file
	f, _ := os.Open(ConfigFilePath(uuidStr))
	fStat, _ := f.Stat()
	fSize := fStat.Size()
	buf := make([]byte, fSize)
	f.Read(buf)
	f.Close()

	return Container{
		Id:         uuidStr,
		ConfigJson: string(buf),
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

		// read config file
		f, _ := os.Open(ConfigFilePath(info.Name()))
		fStat, _ := f.Stat()
		fSize := fStat.Size()
		buf := make([]byte, fSize)
		f.Read(buf)
		f.Close()

		cs = append(cs, Container{
			Id:         info.Name(),
			ConfigJson: string(buf),
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
