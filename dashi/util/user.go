package util

import "os"

func IsRunningRootUser() bool {
	return os.Geteuid() == 0
}
