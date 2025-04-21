package container

import (
	"github.com/sirupsen/logrus"
	"os"
	"syscall"
)

// RunContainerInitProcess 初始化容器，创建 namespace 和 cgroups 资源限制
// 参数 command 是要执行的命令，args 是命令的参数列表
func RunContainerInitProcess(command string, args []string) error {
	logrus.Infof("command %s", command)
	// 设置挂载参数，禁止执行权限、设置用户ID和设备访问
	//MS_NOEXEC在文件系统中不允许运行其他程序, MS_NOSUID在文件系统中不允许set-user-ID或set-group-ID, MS_NODEV是mount系统默认设定的参数
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV

	// 将 proc 文件系统挂载到 /proc
	// proc 文件系统提供了进程和内核的信息，是容器中最基本的虚拟文件系统
	// 例如 ps、top 等命令都依赖 /proc 提供的进程信息
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	// 替换当前进程镜像，执行用户传入的命令（不会返回）
	if err := syscall.Exec(command, args, os.Environ()); err != nil {
		logrus.Errorf("Exec command error %v", err.Error())
	}
	return nil
}
