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
	"path/filepath"
	"strings"
	"syscall"
)

// RunContainerInitProcess 本容器执行的第一个进程
// 使用mount挂载proc文件系统
// 以便后面通过ps等系统命令查看当前进程资源的情况
func RunContainerInitProcess(asChild bool) error {
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

	if asChild {
		err = runProcessAsChild(cmdArray)
	} else {
		err = runProcessInsteadParent(cmdArray)
	}
	if err != nil {
		logrus.Errorf("run container init process failed. commands:%v, err: %v", cmdArray, err)
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

	var err error

	wd, _ := os.Getwd()
	logrus.Infof("set up mount, current location is: %s", wd)

	// systemd加入linux后，mount namespace就变成shared by default，
	// 所以必须显式声明要这个新的mount namespace独立
	err = syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err != nil {
		return err
	}

	err = pivotRoot()
	if err != nil {
		logrus.Errorf("pivot root, err: %v", err)
		return err
	}

	// mount proc
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	if err != nil {
		logrus.Errorf("mount proc, err: %v", err)
		return err
	}

	// mount tmpfs, tmpfs是一种基于内存的文件系统
	err = syscall.Mount("tmpfs", "/dev", "tmpfs",
		syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
	if err != nil {
		logrus.Errorf("mount tmpfs, ere: %v", err)
		return err
	}

	return nil
}

// 改变当前root的文件系统
func pivotRoot() error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	logrus.Infof("pivot root current location is: %s", root)

	// 为了使当前root的老root和新root不在同一个文件系统下，我们把root重新mount了一次
	// bind mount是把相同的内容换了一个挂载点的挂载方法
	err = syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, "")
	if err != nil {
		return fmt.Errorf("mount rootfs to itself error: %v", err)
	}

	// 创建rootfs/.pivot_root 存储 old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	_, err = os.Stat(pivotDir)
	if err != nil && os.IsNotExist(err) {
		if err := os.Mkdir(pivotDir, 0777); err != nil {
			return err
		}
	}

	// pivot_root 到新的rootfs，现在老的 old_root 是挂载在rootfs/.pivot_root
	// 挂载点现在依然可以在mount命令中看到
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}

	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}

	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
	}

	// 删除临时文件夹
	return os.Remove(pivotDir)
}
