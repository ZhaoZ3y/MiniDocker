package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

// MemorySubSystem 是 memory 子系统的实现，用于设置内存限制
type MemorySubSystem struct{}

// Set 设置某个 cgroup 在 memory 子系统中的内存限制
// 参数 cgroupPath 是 cgroup 的相对路径，例如 "mydocker-cgroup/container1"
// 参数 res 包含用户设置的资源限制（这里使用 res.MemoryLimit）
func (s *MemorySubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	// 获取对应子系统的绝对路径，并确保目录存在（create = true）
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		// 如果用户设置了内存限制，就写入到 memory.limit_in_bytes 文件中
		if res.MemoryLimit != "" {
			// 将内存限制写入 memory 子系统的配置文件中（单位是字节）
			if err := ioutil.WriteFile(
				path.Join(subsysCgroupPath, "memory.limit_in_bytes"),
				[]byte(res.MemoryLimit),
				0644); err != nil {
				return fmt.Errorf("设置内存限制失败: %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

// Remove 删除对应 cgroup 在 memory 子系统中的目录
// 这通常在容器退出时调用
func (s *MemorySubSystem) Remove(cgroupPath string) error {
	// 获取对应路径（create = false）
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		// 删除目录（会清理所有限制设置）
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return err
	}
}

// Apply 将某个进程（pid）添加到 memory 子系统的 cgroup 中
func (s *MemorySubSystem) Apply(cgroupPath string, pid int) error {
	// 获取该 cgroup 的路径
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		// 向 tasks 文件写入 pid，即将进程加入到该 cgroup 中
		if err := ioutil.WriteFile(
			path.Join(subsysCgroupPath, "tasks"),
			[]byte(strconv.Itoa(pid)),
			0644); err != nil {
			return fmt.Errorf("加入内存 cgroup 失败: %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("获取内存 cgroup 失败: %v", err)
	}
}

// Name 返回该子系统的名称，用于在 /sys/fs/cgroup 下定位路径
func (s *MemorySubSystem) Name() string {
	return "memory"
}
