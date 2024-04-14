package internal

const CONTAINERS_PATH string = "./containers"
const ROOTFS_DIR_NAME string = "rootfs"

func RootfsPath(cId string) string {
	return CONTAINERS_PATH + "/" + cId + "/" + ROOTFS_DIR_NAME
}
