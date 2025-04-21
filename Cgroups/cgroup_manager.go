package Cgroups

import "MiniDocker/Cgroups/subsystems"

// CgroupManager 用于管理 cgroup 的结构体
type CgroupManager struct {
	// cgroup在hierarchy中的路径 相当于创建的cgroup目录相对于root cgroup目录的路径
	Path string
	// 资源配置
	Resource *subsystems.ResourceConfig
}

// NewCgroupManager 创建一个新的 CgroupManager 实例
func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

// Apply 将进程 pid 加入到这个 cgroup 中
func (c *CgroupManager) Apply(pid int) error {
	for _, subSysIns := range subsystems.SubSystemsIns {
		if err := subSysIns.Apply(c.Path, pid); err != nil {
			return err
		}
	}
	return nil
}

// Set 设置 cgroup 资源限制
func (c *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	for _, subSysIns := range subsystems.SubSystemsIns {
		if err := subSysIns.Set(c.Path, res); err != nil {
			return err
		}
	}
	return nil
}

// Destroy 释放 cgroup
func (c *CgroupManager) Destroy() error {
	for _, subSysIns := range subsystems.SubSystemsIns {
		if err := subSysIns.Remove(c.Path); err != nil {
			return err
		}
	}
	return nil
}
