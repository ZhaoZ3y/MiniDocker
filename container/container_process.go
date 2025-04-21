package container

import (
	"os"
	"os/exec"
	"syscall"
)

// NewParentProcess 创建一个新的父进程（容器的父进程）
// 参数 tty 表示是否需要终端（即是否将容器进程的输入输出绑定到宿主机终端）
// 参数 command 是要在容器中执行的初始命令
func NewParentProcess(tty bool, command string) *exec.Cmd {
	// 构造传递给子进程的参数，第一个参数 "init" 用于标识初始化流程，
	// 第二个是实际要在容器中执行的命令
	args := []string{"init", command}

	// 创建一个执行 /proc/self/exe 的命令，也就是再次执行当前程序自身
	// 子进程启动后会根据 args 执行容器初始化流程
	cmd := exec.Command("/proc/self/exe", args...)

	// 设置子进程的命名空间隔离属性
	// 使用 CLONE_NEWUTS：隔离主机名
	// 使用 CLONE_NEWPID：隔离进程号
	// 使用 CLONE_NEWNS：隔离挂载点
	// 使用 CLONE_NEWNET：隔离网络
	// 使用 CLONE_NEWIPC：隔离进程间通信
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWIPC,
	}

	// 如果 tty 为 true，表示需要与用户交互
	// 将子进程的标准输入、输出和错误输出连接到当前终端
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// 返回构造好的命令对象，由调用者负责启动执行
	return cmd
}
