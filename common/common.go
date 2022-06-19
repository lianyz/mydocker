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