/*
@Time : 2022/6/16 11:31
@Author : lianyz
@Description :
*/

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
)

const cgroupMemoryHierarchyMount = "/sys/fs/cgroup/memory"

const memoryChildCgroup = "testmemorylimit"

func main() {

	if len(os.Args) < 2 {
		fmt.Println("usage: cgroup 100m")
		return
	}

	if os.Args[0] == "/proc/self/exe" {
		fmt.Printf("run %s process......\n", os.Args[0])

		// 容器进程
		fmt.Printf("current pid:%d ppid:%d\n", syscall.Getpid(), syscall.Getppid())
		stress := fmt.Sprintf("stress --vm-bytes %s --vm-keep -m 1", os.Args[1])
		fmt.Println(stress)

		cmd := exec.Command("sh", "-c", stress)
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("cmd [sh -c stress...] pid:%d \n", cmd.Process.Pid)

		cmd.Wait()

		fmt.Println("can not run to here......")

		return
	}

	fmt.Printf("run %s process......\n", os.Args[0])
	fmt.Printf("current pid:%d ppid:%d\n", syscall.Getpid(), syscall.Getppid())

	cmd := exec.Command("/proc/self/exe", os.Args[1])
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS,
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("cmd [/proc/self/exe] pid:%d \n", cmd.Process.Pid)
	fmt.Printf("ProcessID: %v\n", cmd.Process.Pid)
	os.Mkdir(path.Join(cgroupMemoryHierarchyMount, memoryChildCgroup), 0755)
	ioutil.WriteFile(path.Join(cgroupMemoryHierarchyMount, memoryChildCgroup, "tasks"),
		[]byte(strconv.Itoa(cmd.Process.Pid)), 0644)
	ioutil.WriteFile(path.Join(cgroupMemoryHierarchyMount, memoryChildCgroup, "memory.limit_in_bytes"),
		[]byte("100m"), 0644)

	cmd.Wait()
}
