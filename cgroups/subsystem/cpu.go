/*
@Time : 2022/6/16 22:57
@Author : lianyz
@Description :
*/

package subsystem

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type CpuSubSystem struct {
}

func (*CpuSubSystem) Name() string {
	return "cpu"
}

func (m *CpuSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	subsystemCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, true)
	if err != nil {
		logrus.Errorf("get %s path, err: %v", cgroupPath, err)
		return err
	}

	if res.CpuShare != "" {
		// 设置cgroup内存限制
		err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "cpu.shares"),
			[]byte(res.CpuShare), 0644)

		if err != nil {
			return err
		}
	}

	return nil
}

func (m *CpuSubSystem) Remove(cgroupPath string) error {
	subsystemCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, true)
	if err != nil {
		return err
	}

	return os.RemoveAll(subsystemCgroupPath)
}

func (m *CpuSubSystem) Apply(cgroupPath string, pid int) error {
	subsystemCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, true)
	if err != nil {
		return err
	}

	tasksPath := path.Join(subsystemCgroupPath, "tasks")
	err = ioutil.WriteFile(tasksPath, []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		logrus.Errorf("write pid to tasks, path: %s, pid: %d, err: %v",
			tasksPath, pid, err)
		return err
	}

	return nil
}
