/*
@Time : 2022/6/16 22:58
@Author : lianyz
@Description :
*/

package subsystem

// ResourceConfig 资源限制配置
type ResourceConfig struct {
	// 内存限制
	MemoryLimit string
	// CPU时间片权重
	CpuShare string
	// CPU核数
	CpuSet string
}

// Subsystem 将cgroup抽象为path，因为在hierarchy中，cgroup就是虚拟的路径地址
type Subsystem interface {
	// Name 返回subsystem名字，如 cpu,memory
	Name() string

	// Set 设置cgroup在这个subSystem中的资源限制
	Set(cgroupPath string, res *ResourceConfig) error

	// Remove 移除这个cgroup的资源限制
	Remove(cgroupPath string) error

	// Apply 将某个进程添加至cgroup中
	Apply(cgroupPath string, pid int) error
}

var (
	Subsystems = []Subsystem{
		&MemorySubSystem{},
		&CpuSubSystem{},
		&CpuSetSubSystem{},
	}
)
