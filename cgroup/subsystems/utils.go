package subsystems

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

// FindCgroupMountpoint 查找指定子系统的 cgroup 挂载点
// 参数 subsystem：子系统的名称（如 "cpu", "memory", "cpuset" 等）
// 返回值：子系统的挂载点路径，找不到时返回空字符串
func FindCgroupMountpoint(subsystem string) string {
	// 打开 /proc/self/mountinfo 文件，这是 Linux 系统的挂载信息
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer f.Close()

	// 按行读取 mountinfo
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()             // 获取每一行的挂载信息
		fields := strings.Split(txt, " ") // 将每行信息按空格拆分为字段

		// 遍历字段中最后一列，查找包含指定子系统名的挂载项
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				return fields[4] // 返回 cgroup 的挂载点路径
			}
		}
	}

	// 如果有错误发生，返回空字符串
	if err := scanner.Err(); err != nil {
		return ""
	}

	return ""
}

// GetCgroupPath 获取指定子系统的 cgroup 路径
// 参数 subsystem：子系统名称（如 "cpu", "memory", "cpuset"）
// 参数 cgroupPath：子系统下的 cgroup 路径（如 "my_cgroup"）
// 参数 autoCreate：是否自动创建路径（若路径不存在时）
// 返回值：cgroup 的绝对路径，或者错误信息
func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error) {
	// 获取该子系统的挂载点路径
	cgroupRoot := FindCgroupMountpoint(subsystem)

	// 判断 cgroupPath 是否存在
	if _, err := os.Stat(path.Join(cgroupRoot, cgroupPath)); err == nil || (autoCreate && os.IsNotExist(err)) {
		// 如果路径不存在且允许自动创建，则创建该 cgroup 路径
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path.Join(cgroupRoot, cgroupPath), 0755); err == nil {
				// 创建成功，返回路径
			} else {
				return "", fmt.Errorf("创建 cgroup 失败: %v", err)
			}
		}
		// 返回 cgroup 的绝对路径
		return path.Join(cgroupRoot, cgroupPath), nil
	} else {
		// 如果路径已存在或发生其他错误，返回错误信息
		return "", fmt.Errorf("获取 cgroup 路径失败: %v", err)
	}
}
