package subsystems

// ResourceConfig 用于传递资源限制配置
// 用户可以通过该结构体限制容器的 CPU、内存等资源
type ResourceConfig struct {
	MemoryLimit string // 内存限制，例如 "500m"
	CpuShare    string // CPU 使用权重，例如 "1024"
	CpuSet      string // CPU 核绑定，例如 "0-2"、"0,1"
}

// Subsystem 接口，每种 cgroup 子系统（如 memory、cpu、cpuset）都实现这个接口
// 这样就能统一管理不同资源类型
type Subsystem interface {
	Name() string                          // 返回子系统名称，如 "cpu"、"memory"
	Set(path string, res *ResourceConfig) error  // 设置资源限制
	Apply(path string, pid int) error     // 将某个进程加入到这个 cgroup 中
	Remove(path string) error             // 删除这个 cgroup
}

// SubsystemsIns 是各个子系统的注册列表（一个工厂）
// 后续可以通过遍历它来统一设置或清理资源限制
var (
	SubsystemsIns = []Subsystem{
		&CpusetSubSystem{},  // CPU 核心绑定限制
		&MemorySubSystem{},  // 内存限制
		&CpuSubSystem{},     // CPU 权重限制
	}
)
