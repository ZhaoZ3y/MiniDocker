package container

import (
	"github.com/sirupsen/logrus"
	"os"
	"syscall"
)

// RunContainerInitProcess 初始化容器，创建 namespace 和 cgroups 的资源限制环境
// 参数 command 是用户指定要执行的命令（如 /bin/bash）
// 参数 args 是该命令对应的参数列表
func RunContainerInitProcess(command string, args []string) error {
	// 打印出用户指定的命令，用于日志调试
	logrus.Infof("command %s", command)

	// 设置挂载参数（组合多个 mount 标志）
	// MS_NOEXEC：不允许执行文件
	// MS_NOSUID：忽略 set-user-ID 和 set-group-ID
	// MS_NODEV：禁止访问设备文件
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV

	// 将 proc 文件系统挂载到 /proc
	// proc 文件系统提供了进程和内核的信息，是容器中最基本的虚拟文件系统
	// 例如 ps、top 等命令都依赖 /proc 提供的进程信息
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")

	// 使用 syscall.Exec 执行用户指定的命令，替换当前进程镜像
	// 这一步执行后当前进程不会再返回，也就是说容器内将直接运行用户指定的命令
	// os.Environ() 会将当前环境变量传入新进程
	if err := syscall.Exec(command, args, os.Environ()); err != nil {
		// 如果 Exec 执行失败，打印错误日志
		logrus.Errorf("Exec command error %v", err.Error())
	}
	return nil
}