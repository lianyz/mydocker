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
	"strings"
)

// NewWorkSpace 创建容器运行时目录
func NewWorkSpace(volume, containerName, imageName string) error {
	// 创建只读层
	err := createReadOnlyLayer(imageName)
	if err != nil {
		logrus.Errorf("create read only layer, err: %v", err)
		return err
	}

	// 创建读写层
	err = createWriteLayer(containerName)
	if err != nil {
		logrus.Errorf("create write layer, err: %v", err)
		return err
	}

	// 创建挂载点，将只读层和读写层挂载到指定位置
	err = createMountPoint(containerName, imageName)
	if err != nil {
		logrus.Errorf("create mount point, err: %v", err)
		return err
	}

	// 设置宿主机与容器文件映射
	mountVolume(containerName, imageName, volume)

	return nil
}

func createReadOnlyLayer(imageName string) error {
	imagePath := path.Join(common.RootPath, imageName)
	if err := common.Mkdir(imagePath); err != nil {
		logrus.Errorf("mkdir image path, err: %v", err)
		return err
	}

	// 解压 /root/imageName.tar
	imageTarPath := path.Join(common.RootPath, fmt.Sprintf("%s.tar", imageName))
	if _, err := exec.Command("tar", "-xvf", imageTarPath, "-C", imagePath).CombinedOutput(); err != nil {
		logrus.Errorf("tar image tar, path: %s, err: %v", imageTarPath, err)
		return err
	}

	return nil
}

// 创建读写层
func createWriteLayer(containerName string) error {
	writeLayerPath := path.Join(common.RootPath, common.WriteLayer, containerName)
	if err := common.Mkdir(writeLayerPath); err != nil {
		logrus.Errorf("mkdir write layer, err: %v", err)
		return err
	}

	return nil
}

func createMountPoint(containerName, imageName string) error {
	mntPath := path.Join(common.MntPath, containerName)
	if err := common.Mkdir(mntPath); err != nil {
		logrus.Errorf("mkdir mnt path, err: %v", err)
		return err
	}

	// 将宿主机上关于容器的读写层和只读层挂载到 /root/mnt/容器名 里
	writeLayerPath := path.Join(common.RootPath, common.WriteLayer, containerName)
	imagePath := path.Join(common.RootPath, imageName)
	dirs := fmt.Sprintf("dirs=%s:%s", writeLayerPath, imagePath)
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntPath)
	if err := cmd.Run(); err != nil {
		logrus.Errorf("mnt cmd run, err: %v", err)
		return err
	}

	return nil
}

func mountVolume(containerName, imageName, volume string) {
	if volume == "" {
		return
	}

	volumes := strings.Split(volume, ":")
	if len(volumes) != 2 {
		logrus.Errorf("volume parameter input is not correct")
		return
	}

	// 创建宿主机中的文件路径
	hostPath := volumes[0]
	if err := common.Mkdir(hostPath); err != nil {
		logrus.Errorf("make host volume path: %s, err: %v", hostPath, err)
		return
	}

	// 创建容器内挂载点
	containerPath := volumes[1]
	containerVolumePath := path.Join(common.MntPath, containerName, containerPath)
	if err := common.Mkdir(containerVolumePath); err != nil {
		logrus.Errorf("make container volume path: %s, err: %v", containerVolumePath, err)
		return
	}

	// 把宿主机文件目录挂载到容器挂载点中
	dirs := fmt.Sprintf("dirs=%s", hostPath)
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logrus.Errorf("mount volume path, err: %v", err)
	}
}

// DeleteWorkSpace 删除容器工作空间
func DeleteWorkSpace(containerName, volume string) error {
	// 删除容器里volume挂载点的文件系统
	deleteVolume(containerName, volume)

	// 卸载整个容器文件系统的挂载点
	err := unMountPoint(containerName)
	if err != nil {
		return err
	}

	// 删除读写层
	err = deleteWriteLayer(containerName)
	if err != nil {
		return err
	}

	return nil
}

func unMountPoint(containerName string) error {
	mntPath := path.Join(common.MntPath, containerName)
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

func deleteWriteLayer(containerName string) error {
	writeLayerPath := path.Join(common.RootPath, common.WriteLayer, containerName)
	return os.RemoveAll(writeLayerPath)
}

func deleteVolume(containerName, volume string) {
	if volume == "" {
		return
	}

	volumes := strings.Split(volume, ":")
	if len(volumes) != 2 {
		logrus.Errorf("volume parameter input is not correct")
		return
	}

	containerVolumePath := path.Join(
		common.MntPath, common.WriteLayer, containerName, volumes[1])
	if _, err := exec.Command("umount", containerVolumePath).CombinedOutput(); err != nil {
		logrus.Errorf("unmount container volume path, err: %v", err)
	}
}
