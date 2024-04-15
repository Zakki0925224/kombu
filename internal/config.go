package internal

const CONTAINERS_PATH string = "./containers"
const ROOTFS_DIR_NAME string = "rootfs"
const OCI_RUNTIME_BUNDLES_PATH string = "./bundles"
const BUNDLE_CONFIG_FILE_NAME string = "config.json"

func RootfsPath(cId string) string {
	return CONTAINERS_PATH + "/" + cId + "/" + ROOTFS_DIR_NAME
}

func ConfigFilePath(cId string) string {
	return CONTAINERS_PATH + "/" + cId + "/" + BUNDLE_CONFIG_FILE_NAME
}
