/*
@Time : 2022/6/16 22:56
@Author : lianyz
@Description :
*/

package main

import (
	"github.com/lianyz/mydocker/cgroups"
	"github.com/lianyz/mydocker/cgroups/subsystem"
	"github.com/lianyz/mydocker/common"
	"github.com/lianyz/mydocker/container"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

func Run(cmdArray []string, tty bool, asChild bool, res *subsystem.ResourceConfig, volume string) {
	parent, writePipe := container.NewParentProcess(tty, asChild, volume)
	if parent == nil {
		logrus.Errorf("failed to new parent process")
		return
	}
	if err := parent.Start(); err != nil {
		logrus.Errorf("parent process start failed, err: %v", err)
		return
	}

	logrus.Infof("start parent process succeed!")

	// 添加资源限制
	cgroupManager := cgroups.NewCGroupManager("mydocker")
	// 删除资源限制
	defer cgroupManager.Destroy()
	// 设置资源限制
	cgroupManager.Set(res)
	// 将容器进程加入到各个subsystem挂载对应的cgroup中
	cgroupManager.Apply(parent.Process.Pid)

	sendInitCommand(cmdArray, writePipe)

	if tty {
		err := parent.Wait()
		if err != nil {
			logrus.Errorf("parent wait, err: %v", err)
		}

		// 删除容器工作空间
		err = container.DeleteWorkSpace(common.RootPath, common.MntPath, volume)
		if err != nil {
			logrus.Errorf("delete work space, err: %v", err)
		}
	}
}

func sendInitCommand(cmdArray []string, writePipe *os.File) {
	command := strings.Join(cmdArray, " ")
	logrus.Infof("command all is: %s", command)
	_, _ = writePipe.WriteString(command)
	_ = writePipe.Close()
}
