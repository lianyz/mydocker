/*
@Time : 2022/6/16 23:44
@Author : lianyz
@Description :
*/

package container

import (
	"github.com/lianyz/mydocker/common"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path"
	"syscall"
)

// NewParentProcess 创建一个会隔离namespace进程的Command
func NewParentProcess(tty bool, asChild bool,
	volume, containerName, imageName string, envs []string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, _ := os.Pipe()

	// 调用自身，传入init参数，也就是执行initCommand
	args := []string{
		"init",
	}
	if asChild {
		args = append(args, "-ch")
	}
	logrus.Infof("/proc/self/exe args:%v", args)
	cmd := exec.Command("/proc/self/exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWIPC,
	}

	if tty {
		cmd.Stdin = os.Stdout
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		logDir := path.Join(common.DefaultContainerInfoPath, containerName)
		if err := common.Mkdir(logDir); err != nil {
			logrus.Errorf("mkdir container log, err: %v", err)
		}

		logFileName := path.Join(logDir, common.ContainerLogFileName)
		file, err := os.Create(logFileName)
		if err != nil {
			logrus.Errorf("create log file, err: %v", err)
		}
		cmd.Stdout = file
	}

	// 设置额外文件句柄
	cmd.ExtraFiles = []*os.File{
		readPipe,
	}

	// 将/bin添加至环境变量
	envs = append(envs, "PATH=$PATH:/bin:/usr/bin")

	// 设置环境变量
	cmd.Env = append(os.Environ(), envs...)

	err := NewWorkSpace(volume, containerName, imageName)
	if err != nil {
		logrus.Errorf("new work space, err: %v", err)
	}

	logrus.Infof("change workdir, new:%s, old:%s", common.MntPath, cmd.Dir)

	// 指定容器初始化后的工作目录
	cmd.Dir = path.Join(common.MntPath, containerName)

	return cmd, writePipe
}
