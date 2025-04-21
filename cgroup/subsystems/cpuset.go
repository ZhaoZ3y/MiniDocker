package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

// CpusetSubSystem 是 cpuset 子系统的实现
// 用于限制容器使用指定的 CPU 核心（如只使用 CPU 0 和 1）
type CpusetSubSystem struct{}

// Set 设置某个 cgroup 在 cpuset 子系统中的资源限制
// 通过写入 cpuset.cpus 文件，指定可以使用的 CPU 编号
// 比如 "0", "0-2", "0,1" 等格式
func (s *CpusetSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	// 获取当前子系统对应的 cgroup 路径（create = true 表示创建路径）
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		// 若用户配置了 CPU 核分配，则写入 cpuset.cpus 文件
		if res.CpuSet != "" {
			if err := ioutil.WriteFile(
				path.Join(subsysCgroupPath, "cpuset.cpus"),
				[]byte(res.CpuSet),
				0644); err != nil {
				return fmt.Errorf("设置 cpuset 失败: %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

// Remove 删除 cpuset 子系统下的 cgroup 目录
func (s *CpusetSubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return err
	}
}

// Apply 将进程 pid 添加到该 cpuset cgroup 中，使其受限于设定的 CPU 核
func (s *CpusetSubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		if err := ioutil.WriteFile(
			path.Join(subsysCgroupPath, "tasks"),
			[]byte(strconv.Itoa(pid)),
			0644); err != nil {
			return fmt.Errorf("添加进程到 cpuset cgroup 失败: %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("获取 cpuset cgroup 失败 %s: %v", cgroupPath, err)
	}
}

// Name 返回当前子系统的名称
func (s *CpusetSubSystem) Name() string {
	return "cpuset"
}
