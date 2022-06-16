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
