package container

import (
	"os"
	"os/exec"
	"syscall"
)

// NewParentProcess 创建一个新的父进程（容器的父进程）
// 参数 tty 表示是否需要终端（是否要把标准输入输出绑定到当前终端）
// 参数 command 是要在容器中执行的命令
func NewParentProcess(tty bool, command string) *exec.Cmd {
	args := []string{"init", command}

	// 创建一个新的命令行进程
	// /proc/self/exe 是当前进程的可执行文件路径
	cmd := exec.Command("/proc/self/exe", args...)
	// 创建新的UTS\PID\IPC\NET\NS命名空间，实现容器隔离
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	// 如果需要终端，则绑定标准输入输出到当前终端，实现交互式命令输入
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd
}
