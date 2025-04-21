package subsystems

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

// FindCgroupMountpoint 查找指定子系统的 cgroup 挂载点所在的目录
func FindCgroupMountpoint(subsystem string) (string, error) {
	// 打开 /proc/self/mountinfo 文件，该文件包含了当前进程的挂载信息
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}
	defer f.Close()

	// 逐行读取文件内容
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// 读取一行文本
		txt := scanner.Text()
		// 将文本按空格分割成字段
		fields := strings.Fields(txt)
		// 遍历最后一个字段（即挂载选项），按逗号分割
		// 每个选项都与传入的子系统名称进行比较
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				// 返回该子系统的 cgroup 挂载点路径
				return fields[4], nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", os.ErrNotExist
}

// GetCgroupPath 获取指定子系统的 cgroup 绝对路径
// cgroupPath: cgroup 的名称
// autoCreate: 是否自动创建 cgroup
func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error) {
	// 查找指定子系统的 cgroup 挂载点所在的目录
	cgroupRoot, err := FindCgroupMountpoint(subsystem)
	if err != nil {
		return "", err
	}
	// 检查 cgroup 路径是否存在
	if _, err := os.Stat(path.Join(cgroupRoot, cgroupPath)); err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			// 如果不存在且需要自动创建，则创建该 cgroup 路径
			if err := os.Mkdir(path.Join(cgroupRoot, cgroupPath), 0755); err != nil {
				return "", fmt.Errorf("创建cgroup路径失败 %v", err)
			}
		}
		return path.Join(cgroupRoot, cgroupPath), nil
	} else {
		return "", fmt.Errorf("cgroup路径错误 %v", err)
	}
}
