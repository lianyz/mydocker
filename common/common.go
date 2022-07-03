/*
@Time : 2022/6/18 18:33
@Author : lianyz
@Description :
*/

package common

import "os"

const (
	RootPath   = "/root/"
	MntPath    = "/root/mnt/"
	BinPath    = "/bin/"
	BusyBox    = "busybox"
	BusyBoxTar = "busybox.tar"
	WriteLayer = "writeLayer"
)

const (
	Running = "running"
	Stop    = "stopped"
	Exit    = "exited"
)

const (
	DefaultContainerInfoPath = "/var/run/mydocker"
	ContainerInfoFileName    = "config.json"
	ContainerLogFileName     = "container.log"
)

const (
	EnvExecPid = "docker_pid"
	EnvExecCmd = "docker_cmd"
)

const (
	DefaultNetworkPath   = "/var/run/mydocker/network/network/"
	DefaultAllocatorPath = "/var/run/mydocker/network/ipam/subnet.json"
)

func Mkdir(path string) error {
	if IsNotExist(path) {
		return os.MkdirAll(path, os.ModePerm)
	}

	return nil
}

func IsNotExist(name string) bool {
	_, err := os.Stat(name)
	return err != nil && os.IsNotExist(err)
}
