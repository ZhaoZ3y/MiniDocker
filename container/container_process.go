package container

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
)

// NewParentProcess 创建一个新的父进程（容器的父进程）
// tty 表示是否启用终端（比如交互式容器就需要）
// 返回值包括：创建的 cmd 命令对象 和 写入端 writePipe，用于父子进程通信
func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {
	// 创建匿名管道：用于父子进程之间通信（传参数或控制信号）
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.Errorf("管道创建失败: %v", err)
		return nil, nil
	}

	// 获取当前程序的可执行路径，准备以 "init" 子命令重新执行自己（类似 fork/exec 模式）
	selfPath, err := os.Executable()
	if err != nil {
		log.Fatalf("获取自身路径失败: %v", err)
	}

	cmd := exec.Command(selfPath, "init") // 子进程会执行 "init" 分支逻辑

	// 设置命名空间隔离（关键点：实现容器隔离）
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | // UTS：主机名隔离
			syscall.CLONE_NEWPID | // PID：进程号隔离
			syscall.CLONE_NEWNS | // Mount：挂载点隔离
			syscall.CLONE_NEWNET | // 网络隔离
			syscall.CLONE_NEWIPC, // IPC 隔离
	}

	// 如果 tty 为 true，就把子进程的标准输入输出指向当前终端
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// 把管道的读端传递给子进程（子进程从这里读取父进程传过来的数据）
	cmd.ExtraFiles = []*os.File{readPipe}

	// 初始化容器的挂载点（写层 + 只读层 + aufs 挂载）
	mountURL := "/root/mnt"
	rootURL := "/root/"
	NewWorkSpace(rootURL, mountURL)

	// 设置子进程的当前工作目录为挂载点目录
	cmd.Dir = mountURL

	// TODO: 可选地设置 cmd.Env = os.Environ() + 额外变量，传递环境变量

	return cmd, writePipe
}

// NewPipe 创建一个匿名管道：用于父子进程之间的通信
// 返回：管道读端、写端
func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}
