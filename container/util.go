/*
@Time : 2022/6/18 10:06
@Author : lianyz
@Description :
*/

package container

import (
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
)

// runProcessAsChild 运行进行，作为父进程的子进程
func runProcessAsChild(cmdArray []string) error {
	cmd := exec.Command(cmdArray[0], cmdArray[1:]...)
	logrus.Infof("command name: %s args: %v", cmdArray[0], cmdArray[1:])
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// runProcessInsteadParent 运行进程，并替换掉父进程
func runProcessInsteadParent(cmdArray []string) error {
	// 在系统环境PATH中寻找命令的绝对路径
	logrus.Infof("begin syscall exec cmds: %v", cmdArray)

	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		logrus.Errorf("look %s path, err: %v", cmdArray[0], err)
		return err
	}

	logrus.Infof("end syscall exec cmds: %v", cmdArray)
	return syscall.Exec(path, cmdArray[0:], os.Environ())
}

func execCommand(cmdName, params string) error {
	cmd := exec.Command(cmdName, params)
	output, err := cmd.Output()
	if err != nil {
		logrus.Errorf("iptables output: %v, err: %v", output, err)
		return err
	}
	return nil
}
