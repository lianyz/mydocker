/*
@Time : 2022/6/18 18:40
@Author : lianyz
@Description :
*/

package container

import (
	"fmt"
	"github.com/lianyz/mydocker/common"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path"
)

// NewWorkSpace 创建容器运行时目录
func NewWorkSpace(rootPath string, mntPath string, volume string) error {
	// 创建只读层
	err := createReadOnlyLayer(rootPath)
	if err != nil {
		logrus.Errorf("create read only layer, err: %v", err)
		return err
	}

	// 创建读写层
	err = createWriteLayer(rootPath)
	if err != nil {
		logrus.Errorf("create write layer, err: %v", err)
		return err
	}

	// 创建挂载点，将只读层和读写层挂载到指定位置
	err = createMountPoint(rootPath, mntPath)
	if err != nil {
		logrus.Errorf("create mount point, err: %v", err)
		return err
	}

	wd, _ := os.Getwd()
	logrus.Infof("create mount point, mntPath:%s current location is: %s", mntPath, wd)

	// 设置宿主机与容器文件映射
	mountVolume(rootPath, mntPath, volume)

	return nil
}

func createReadOnlyLayer(rootPath string) error {
	busyBoxPath := path.Join(rootPath, common.BusyBox)
	_, err := os.Stat(busyBoxPath)
	if err != nil && os.IsNotExist(err) {
		err := os.MkdirAll(busyBoxPath, os.ModePerm)
		if err != nil {
			logrus.Errorf("mkdir busybox, err: %v", err)
			return err
		}
	}

	// 解压 busybox.tar
	busyBoxTarPath := path.Join(rootPath, common.BusyBoxTar)
	if _, err = exec.Command("tar", "-xvf", busyBoxTarPath, "-C", busyBoxPath).CombinedOutput(); err != nil {
		logrus.Errorf("tar busybox.tar, err: %v", err)
		return err
	}

	return nil
}

func createWriteLayer(rootPath string) error {
	writeLayerPath := path.Join(rootPath, common.WriteLayer)
	_, err := os.Stat(writeLayerPath)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(writeLayerPath, os.ModePerm)
		if err != nil {
			logrus.Errorf("mkdir write layer, err: %v", err)
			return err
		}
	}
	return nil
}

func createMountPoint(rootPath string, mntPath string) error {
	_, err := os.Stat(mntPath)
	if err != nil && os.IsNotExist(err) {
		err := os.MkdirAll(mntPath, os.ModePerm)
		if err != nil {
			logrus.Errorf("mkdir mnt path, err: %v", err)
			return err
		}
	}

	dirs := fmt.Sprintf("dirs=%s%s:%s%s", rootPath, common.WriteLayer, rootPath, common.BusyBox)
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntPath)
	if err := cmd.Run(); err != nil {
		logrus.Errorf("mnt cmd run, err: %v", err)
		return err
	}

	return nil
}

func mountVolume(rootPath, mntPath, volume string) {

}

// DeleteWorkSpace 删除容器工作空间
func DeleteWorkSpace(rootPath, mntPath, volume string) error {
	// 卸载挂载点
	err := unMountPoint(mntPath)
	if err != nil {
		return err
	}

	// 删除读写层
	err = deleteWriteLayer(rootPath)
	if err != nil {
		return err
	}

	// 删除宿主机与文件系统映射
	deleteVolume(mntPath, volume)

	return nil
}

func unMountPoint(mntPath string) error {
	if _, err := exec.Command("umount", mntPath).CombinedOutput(); err != nil {
		logrus.Errorf("unmount mnt, err: %v", err)
		return err
	}

	err := os.RemoveAll(mntPath)
	if err != nil {
		logrus.Errorf("remove mnt path, err: %v", err)
		return err
	}

	return nil
}

func deleteWriteLayer(rootPath string) error {
	writeLayerPath := path.Join(rootPath, common.WriteLayer)
	return os.RemoveAll(writeLayerPath)
}

func deleteVolume(mntPath, volume string) {

}
