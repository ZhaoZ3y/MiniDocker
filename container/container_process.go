package container

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
)

// NewParentProcess 创建一个新的父进程（容器的父进程）
// 参数 tty 表示是否需要终端（是否要把标准输入输出绑定到当前终端）
// 返回值是新创建的命令对象（*exec.Cmd）和一个写端管道（*os.File），用于向子进程传递用户命令
func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {
	// 创建匿名管道，用于父进程与子进程通信（这里主要用来传输用户命令）
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.Errorf("创建管道失败 %v", err)
		return nil, nil
	}

	// 创建一个新的命令行进程
	// /proc/self/exe 是当前进程的可执行文件路径, 参数 "init" 表示调用 init 子命令（对应 cli 中的 initCommand）
	cmd := exec.Command("/proc/self/exe", "init")
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
	// 设置管道的读端为命令的标准输入
	// 将读取命令的管道（readPipe）传递给子进程，作为额外的文件描述符
	// 容器中的 init 进程会通过这个文件描述符来读取用户要执行的命令
	cmd.ExtraFiles = []*os.File{readPipe}
	return cmd, writePipe
}

// NewPipe 创建一个新的管道，用于进程间通信
// 返回读端和写端的文件描述符
func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}
