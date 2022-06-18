/*
@Time : 2022/6/16 23:42
@Author : lianyz
@Description :
*/

package container

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// RunContainerInitProcess 本容器执行的第一个进程
// 使用mount挂载proc文件系统
// 以便后面通过ps等系统命令查看当前进程资源的情况
func RunContainerInitProcess() error {
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("get user command in run container")
	}

	// 挂载
	err := setUpMount()
	if err != nil {
		logrus.Errorf("set up mount, err: %v", err)
		return err
	}

	// 在系统环境PATH中寻找命令的绝对路径
	//path, err := exec.LookPath(cmdArray[0])
	//if err != nil {
	//	logrus.Errorf("look %s path, err: %v", cmdArray[0], err)
	//	return err
	//}

	command := strings.Join(cmdArray[1:], " ")
	cmd := exec.Command(cmdArray[0], command)
	logrus.Infof("command name: %s args: %s", cmdArray[0], command)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	//err = syscall.Exec(path, cmdArray[0:], os.Environ())
	if err != nil {
		return err
	}

	return nil
}

func readUserCommand() []string {
	// 指index为3的文件描述符
	// 也就是 cmd.ExtraFiles 中传递过来的 readPipe
	pipe := os.NewFile(uintptr(3), "pipe")
	bs, err := ioutil.ReadAll(pipe)
	if err != nil {
		logrus.Errorf("read pipe, err: %v", err)
		return nil
	}

	msg := string(bs)
	return strings.Split(msg, " ")
}

func setUpMount() error {
	// systemd加入linux后，mount namespace就变成shared by default，
	// 所以必须显式声明要这个新的mount namespace独立
	err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err != nil {
		return err
	}

	// mount proc
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	if err != nil {
		logrus.Errorf("mount proc, err: %v", err)
		return err
	}

	return nil
}
