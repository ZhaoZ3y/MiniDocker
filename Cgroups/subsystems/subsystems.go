package subsystems

// ResourceConfig 资源限制配置，用于传递给 cgroup 子系统
type ResourceConfig struct {
	// MemoryLimit 内存限制
	MemoryLimit string
	// CPUShares CPU 权重
	CPUShares string
	// CPUSet CPU 核心限制
	CPUSet string
}

// Subsystem 接口定义了 cgroup 子系统的基本操作
// 主要包括设置资源限制、将进程添加到 cgroup 中和删除 cgroup
// 将cgroup抽象成了path，原因是cgroup在hierarchy的路径，便是虚拟文件系统的虚拟路径
type Subsystem interface {
	// Name 返回子系统名称
	Name() string
	// Set 设置 cgroup 的资源限制
	Set(path string, res *ResourceConfig) error
	// Apply 将进程添加到 cgroup 中
	Apply(path string, pid int) error
	// Remove 删除 cgroup
	Remove(path string) error
}

// SubSystemsIns 是一个包含所有子系统实例的切片，初始化实例创建资源限制处理链数组
var SubSystemsIns = []Subsystem{
	&MemorySubSystem{},
	&CPUSubSystem{},
	&CPUSetSubSystem{},
}
