package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

// CpuSubSystem 是 cpu 子系统的实现，用于设置 CPU 使用权重（shares）
type CpuSubSystem struct{}

// Set 设置某个 cgroup 在 cpu 子系统中的资源限制
// 这里通过设置 cpu.shares 来限制进程的 CPU 调度权重
// 权重越高，进程分配到的 CPU 时间越多（相对的）
func (s *CpuSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	// 获取当前子系统对应的 cgroup 路径（create = true 会创建目录）
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		// 如果配置了 CPU 权重（如 "512"），写入 cpu.shares 文件
		if res.CpuShare != "" {
			if err := ioutil.WriteFile(
				path.Join(subsysCgroupPath, "cpu.shares"),
				[]byte(res.CpuShare),
				0644); err != nil {
				return fmt.Errorf("设置 CPU 权重失败: %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

// Remove 删除对应 cgroup 在 cpu 子系统中的目录
// 主要用于容器资源回收
func (s *CpuSubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return err
	}
}

// Apply 将进程 pid 加入到该 cgroup 中，使其受到 CPU 限制
func (s *CpuSubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		if err := ioutil.WriteFile(
			path.Join(subsysCgroupPath, "tasks"),
			[]byte(strconv.Itoa(pid)),
			0644); err != nil {
			return fmt.Errorf("将进程加入 CPU cgroup 失败: %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("获取 CPU cgroup 失败 %s: %v", cgroupPath, err)
	}
}

// Name 返回该子系统的名称，用于路径拼接
func (s *CpuSubSystem) Name() string {
	return "cpu"
}
