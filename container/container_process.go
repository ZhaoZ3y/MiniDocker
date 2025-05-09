package container

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
)

// 定义容器状态常量
var (
	RUNNING             string = "running"                 // 容器运行中
	STOPPED             string = "stopped"                 // 容器已停止
	EXIT                string = "exit"                    // 容器已退出
	DefaultInfoLocation string = "/var/run/MiniDocker/%s/" // 容器信息存储路径
	ConfigName          string = "config.json"             // 容器配置文件名
	ContainerLogFile    string = "container.log"
	RootURL             string = "/root"               // 容器根目录
	MntURL              string = "/root/mnt/%s"        // 容器挂载点目录
	WriteLayerURL       string = "/root/writeLayer/%s" // 容器写层目录
)

// Info 结构体定义了容器的基本信息
// 包括 PID、ID、名称、命令、创建时间和状态等字段
type Info struct {
	Pid         string   `json:"pid"`         // 容器的 init 进程在宿主机上的 PID
	Id          string   `json:"id"`          // 容器 ID
	Name        string   `json:"name"`        // 容器名
	Command     string   `json:"command"`     // 容器内 init 运行命令
	CreatedTime string   `json:"createTime"`  // 创建时间
	Status      string   `json:"status"`      // 容器的状态
	Volume      string   `json:"volume"`      // 容器的数据卷
	PortMapping []string `json:"portMapping"` // 容器的端口映射
}

// NewParentProcess 创建一个新的父进程（容器的父进程）
// tty 表示是否启用终端（比如交互式容器就需要）
// 返回值包括：创建的 cmd 命令对象 和 写入端 writePipe，用于父子进程通信
func NewParentProcess(tty bool, volume string, containerName string, imageName string, envSlice []string) (*exec.Cmd, *os.File) {
	// 创建匿名管道：用于父子进程之间通信（传参数或控制信号）
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		logrus.Errorf("管道创建失败: %v", err)
		return nil, nil
	}

	// 获取当前进程的路径（init 进程）
	cmd := exec.Command("/proc/self/exe", "init")

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
	} else {
		// 创建日志目录
		dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
		if err := os.MkdirAll(dirURL, 0622); err != nil {
			logrus.Errorf("创建目录 %v 失败: %v", dirURL, err)
		}

		// 创建日志文件
		stdLogFilePath := dirURL + ContainerLogFile
		stdLogFile, err := os.Create(stdLogFilePath)
		if err != nil {
			logrus.Errorf("创建日志文件 %v 失败: %v", stdLogFilePath, err)
			return nil, nil
		}

		// 设置子进程的标准输出和错误输出为日志文件
		cmd.Stdout = stdLogFile
		cmd.Stderr = stdLogFile

		// 为非交互式命令提供一个空的输入源
		devNull, err := os.Open("/dev/null")
		if err != nil {
			logrus.Errorf("打开 /dev/null 失败: %v", err)
			return nil, nil
		}
		cmd.Stdin = devNull
	}

	// 把管道的读端传递给子进程（子进程从这里读取父进程传过来的数据）
	cmd.ExtraFiles = []*os.File{readPipe}
	cmd.Env = append(os.Environ(), envSlice...)    // 设置环境变量
	NewWorkSpace(volume, imageName, containerName) // 创建工作空间
	logrus.Infof("传递给容器的环境变量: %v", envSlice)
	// 设置子进程的当前工作目录为挂载点目录
	cmd.Dir = fmt.Sprintf(MntURL, containerName)
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
