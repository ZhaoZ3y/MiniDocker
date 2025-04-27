package cgroup

import (
	"MiniDocker/cgroup/subsystems"
	"MiniDocker/container"
	"github.com/sirupsen/logrus"
)

// CgroupManager 管理cgroup的结构体
type CgroupManager struct {
	// cgroup在hierarchy中的路径 相当于创建的cgroup目录相对于root cgroup目录的路径
	Path string
	// 资源配置
	Resource *subsystems.ResourceConfig
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

// Apply 将进程pid加入到这个cgroup中
func (c *CgroupManager) Apply(pid int) error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		subSysIns.Apply(c.Path, pid)
	}
	return nil
}

// Set 设置cgroup资源限制
func (c *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	for _, subSysIns := range subsystems.SubsystemsIns {
		subSysIns.Set(c.Path, res)
	}
	return nil
}

// Destroy 释放cgroup
func (c *CgroupManager) Destroy() error {
	// 先检查 cgroup 是否存在
	if exists, _ := container.PathExists(c.Path); !exists {
		logrus.Warnf("cgroup %s 不存在，跳过删除", c.Path)
		return nil
	}
	for _, subSysIns := range subsystems.SubsystemsIns {
		if err := subSysIns.Remove(c.Path); err != nil {
			logrus.Warnf("删除cgroup %s 失败: %v", c.Path, err)
		}
	}
	return nil
}
