package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

// CPUSubSystem CPU 子系统
type CPUSubSystem struct {
}

// Set 设置内存限制
func (s *CPUSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	// 获取子系统的 cgroup 路径
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.MemoryLimit != "" {
			// 设置内存限制，写入到 memory.limit_in_bytes 文件
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "memory.limit_in_bytes"), []byte(res.MemoryLimit), 0644); err != nil {
				return fmt.Errorf("set cgroup memory limit fail %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

// Remove 删除内存 cgroup
func (s *CPUSubSystem) Remove(cgroupPath string) error {
	// 获取子系统的 cgroup 路径
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return err
	}
}

// Apply 将进程加入到内存 cgroup
func (s *CPUSubSystem) Apply(cgroupPath string, pid int) error {
	// 获取子系统的 cgroup 路径
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		// 将进程 PID 写入到 tasks 文件中
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}
}

// Name 返回内存子系统的名称
func (s *CPUSubSystem) Name() string {
	return "memory"
}
