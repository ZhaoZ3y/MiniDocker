package container

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
)

// NewParentProcess 创建一个新的父进程（容器的父进程）
// 参数 tty 表示是否需要终端（是否要把标准输入输出绑定到当前终端）
// 返回值包括：创建的命令对象 cmd，以及写入端管道 writePipe（用于传递参数给子进程）
func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {
	// 创建一个无名管道，用于父子进程通信
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.Errorf("管道创建失败: %v", err)
		return nil, nil
	}

	// 创建一个新的命令，该命令会重新执行当前程序，并传递参数 "init"
	// init 子命令负责容器中的初始化操作
	cmd := exec.Command("/proc/self/exe", "init")

	// 设置命令的系统属性，使其运行在新的命名空间中，实现隔离
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | // 主机名隔离
			syscall.CLONE_NEWPID | // 进程号隔离
			syscall.CLONE_NEWNS | // 挂载点隔离
			syscall.CLONE_NEWNET | // 网络隔离
			syscall.CLONE_NEWIPC, // 进程间通信隔离
	}

	// 如果需要终端交互，将标准输入输出连接到宿主机的终端
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// 将管道读端传递给子进程（作为 ExtraFile），用于父子通信
	cmd.ExtraFiles = []*os.File{readPipe}

	return cmd, writePipe
}

// NewPipe 创建一个无名管道，用于父子进程之间的通信
// 返回值包括：读端、写端、错误信息
func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}
