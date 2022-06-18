/*
@Time : 2022/6/16 23:44
@Author : lianyz
@Description :
*/

package container

import (
	"os"
	"os/exec"
	"syscall"
)

// NewParentProcess 创建一个会隔离namespace进程的Command
func NewParentProcess(tty bool, asChild bool) (*exec.Cmd, *os.File) {
	readPipe, writePipe, _ := os.Pipe()

	// 调用自身，传入init参数，也就是执行initCommand
	args := "init"
	if asChild {
		args += " -ch"
	}
	cmd := exec.Command("/proc/self/exe", args)
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
	}

	cmd.ExtraFiles = []*os.File{
		readPipe,
	}

	return cmd, writePipe
}
