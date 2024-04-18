package internal

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	specs_go "github.com/opencontainers/runtime-spec/specs-go"
	cp "github.com/otiai10/copy"
)

type Container struct {
	Id   string
	Spec specs_go.Spec
}

func NewContainer(ociRuntimeBundlePath string) (Container, error) {
	uuidObj, err := uuid.NewRandom()
	if err != nil {
		return Container{}, err
	}

	uuidStr := uuidObj.String()

	// create container and rootfs directory
	if err := os.MkdirAll(RootfsPath(uuidStr), os.ModePerm); err != nil {
		return Container{}, err
	}
	// copy rootfs
	if err := cp.Copy(ociRuntimeBundlePath+"/"+ROOTFS_DIR_NAME, RootfsPath(uuidStr)); err != nil {
		return Container{}, err
	}
	// copy config.json
	if err := cp.Copy(ociRuntimeBundlePath+"/"+BUNDLE_CONFIG_FILE_NAME, ConfigFilePath(uuidStr)); err != nil {
		return Container{}, err
	}
	// read config file
	f, err := os.Open(ConfigFilePath(uuidStr))
	if err != nil {
		return Container{}, err
	}

	fStat, err := f.Stat()
	if err != nil {
		return Container{}, err
	}

	fSize := fStat.Size()
	buf := make([]byte, fSize)
	if _, err := f.Read(buf); err != nil {
		return Container{}, err
	}
	if err := f.Close(); err != nil {
		return Container{}, err
	}

	var spec specs_go.Spec
	if err := json.Unmarshal(buf, &spec); err != nil {
		return Container{}, err
	}

	return Container{
		Id:   uuidStr,
		Spec: spec,
	}, nil
}

func FindContainersFromDirectory() ([]Container, error) {
	cs := make([]Container, 0)

	if _, err := os.Stat(CONTAINERS_PATH); os.IsNotExist(err) {
		return cs, err
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
		f, err := os.Open(ConfigFilePath(info.Name()))
		if err != nil {
			return err
		}

		fStat, err := f.Stat()
		if err != nil {
			return err
		}

		fSize := fStat.Size()
		buf := make([]byte, fSize)
		if _, err := f.Read(buf); err != nil {
			return err
		}

		if err := f.Close(); err != nil {
			return err
		}

		var spec specs_go.Spec
		if err := json.Unmarshal(buf, &spec); err != nil {
			return err
		}

		cs = append(cs, Container{
			Id:   info.Name(),
			Spec: spec,
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
