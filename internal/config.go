package internal

const CONTAINERS_PATH string = "./containers"
const OCI_RUNTIME_BUNDLES_PATH string = "./bundles"
const BUNDLE_CONFIG_FILE_NAME string = "config.json"

func ContainerPath(cId string) string {
	return CONTAINERS_PATH + "/" + cId
}

func RootfsPath(cId string, dirPath string) string {
	return ContainerPath(cId) + "/" + dirPath
}

func ConfigFilePath(cId string) string {
	return ContainerPath(cId) + "/" + BUNDLE_CONFIG_FILE_NAME
}
