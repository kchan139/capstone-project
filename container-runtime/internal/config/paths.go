package config

const ConfigURLTemplate = "https://raw.githubusercontent.com/kchan139/capstone-project/main/container-runtime/configs/examples/ubuntu.json"

// Default base directory for container images
const BaseImageDir = "/var/lib/mrunc/images"

// Distro subpaths
func UbuntuRootFS() string {
	return BaseImageDir + "/ubuntu"
}

// func AlpineRootFS() string {
//     return BaseImageDir + "/alpine"
// }
