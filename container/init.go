package container

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// RunContainerInitProcess 是容器中的第一个进程（PID 为 1）
// 它的任务是：
// 1. 设置挂载点（比如挂载 /proc）
// 2. 读取用户要运行的命令
// 3. 使用 syscall.Exec 执行这个命令，替换 init 进程本身
func RunContainerInitProcess() error {
	// 从管道中读取用户传入的命令（通过 ExtraFiles fd[3] 传入）
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("容器初始化获取用户命令错误，cmdArray 为空")
	}

	// 设置默认挂载参数：
	// MS_NOEXEC：不允许执行二进制
	// MS_NOSUID：不允许 set-user-ID 或 set-group-ID
	// MS_NODEV：不允许访问设备文件
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV

	// 挂载 proc 文件系统到 /proc 目录
	// /proc 提供了内核和进程相关的虚拟信息，容器中 ps/top 等命令依赖它
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")

	// 查找要执行命令的绝对路径
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		logrus.Errorf("查找路径失败: %v", err)
		return err
	}
	logrus.Infof("找到可执行文件路径: %s", path)

	// 使用 syscall.Exec 替换当前进程为用户指定的命令进程
	// cmdArray 是命令及其参数，os.Environ() 传入当前环境变量
	if err := syscall.Exec(path, cmdArray, os.Environ()); err != nil {
		logrus.Errorf("执行用户命令失败: %v", err)
	}
	return nil
}

// readUserCommand 通过文件描述符 3（fd3）读取用户传入的命令字符串
// 父进程通过管道写入，子进程通过 fd=3 的文件描述符读取
func readUserCommand() []string {
	// 注意：fd=3 是通过 ExtraFiles 传入的管道读端
	pipe := os.NewFile(uintptr(3), "pipe")
	defer pipe.Close()

	// 读取完整命令字符串
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		logrus.Errorf("从管道读取命令失败: %v", err)
		return nil
	}

	// 将命令字符串按空格分割为命令数组
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}
