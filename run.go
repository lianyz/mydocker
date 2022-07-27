/*
@Time : 2022/6/16 22:56
@Author : lianyz
@Description :
*/

package main

import (
	"github.com/lianyz/mydocker/cgroups"
	"github.com/lianyz/mydocker/cgroups/subsystem"
	"github.com/lianyz/mydocker/container"
	"github.com/lianyz/mydocker/network"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
)

func Run(cmdArray []string, tty bool, asChild bool, res *subsystem.ResourceConfig,
	volume, containerName, imageName, net string,
	envs, ports []string) {
	parent, writePipe := container.NewParentProcess(tty, asChild, volume, containerName, imageName, envs)
	if parent == nil {
		logrus.Errorf("failed to new parent process")
		return
	}
	if err := parent.Start(); err != nil {
		logrus.Errorf("parent process start failed, err: %v", err)
		return
	}

	logrus.Infof("start parent process succeed. pid:%d", parent.Process.Pid)

	containerID := container.GenContainerID(10)
	if containerName == "" {
		containerName = containerID
	}
	// 记录容器信息
	err := container.RecordContainerInfo(parent.Process.Pid, cmdArray, containerName, containerID)
	if err != nil {
		logrus.Errorf("record container info, err: %v", err)
	}
	// 添加资源限制
	cgroupManager := cgroups.NewCGroupManager("mydocker")

	logrus.Infof("run set process resource limit")

	// 设置资源限制
	cgroupManager.Set(res)
	// 将容器进程加入到各个subsystem挂载对应的cgroup中
	cgroupManager.Apply(parent.Process.Pid)

	// 设置网络
	containerInfo := &container.ContainerInfo{
		Id:          containerID,
		Pid:         strconv.Itoa(parent.Process.Pid),
		Name:        containerName,
		PortMapping: ports,
	}
	setNetwork(net, containerInfo)

	// 设置初始化命令
	sendInitCommand(cmdArray, writePipe)

	if tty {
		err := parent.Wait()
		if err != nil {
			logrus.Errorf("parent wait, err: %v", err)
		}

		// 删除容器工作空间
		err = container.DeleteWorkSpace(containerName, volume)
		if err != nil {
			logrus.Errorf("delete work space, err: %v", err)
		}

		logrus.Infof("run begin destroy cgroup resource limit")

		// 删除资源限制
		cgroupManager.Destroy()

		// 删除容器信息
		container.DeleteContainerInfo(containerName)
	}

	logrus.Infof("run finished")
}

func setNetwork(net string, containerInfo *container.ContainerInfo) {
	if net == "" {
		return
	}

	err := network.Init()
	if err != nil {
		logrus.Errorf("network init failed, err: %v", err)
		return
	}

	if err := network.Connect(net, containerInfo); err != nil {
		logrus.Errorf("connect network, err: %v", err)
		return
	}
}

func sendInitCommand(cmdArray []string, writePipe *os.File) {
	command := strings.Join(cmdArray, " ")
	logrus.Infof("command all is: %s", command)
	_, err := writePipe.WriteString(command)
	if err != nil {
		logrus.Errorf("send init command write pipe, err: %v", err)
		return
	}
	if err := writePipe.Close(); err != nil {
		logrus.Errorf("send init command close pipe, err: %v", err)
	}
}
