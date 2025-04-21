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

// RunContainerInitProcess 初始化容器，创建 namespace 和 cgroups 资源限制
// 参数 command 是要执行的命令，args 是命令的参数列表
func RunContainerInitProcess() error {
	// 从管道中读取用户命令
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("获取用户命令失败，cmdArray 为空")
	}
	// 设置挂载参数，禁止执行权限、设置用户ID和设备访问
	//MS_NOEXEC在文件系统中不允许运行其他程序, MS_NOSUID在文件系统中不允许set-user-ID或set-group-ID, MS_NODEV是mount系统默认设定的参数
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV

	// 挂载 proc 文件系统到 /proc，proc 提供了当前系统的进程信息
	// 这是容器中最基础的文件系统，运行 ps 等命令需要用到
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	// path 是用户传入的命令的路径
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		logrus.Errorf("查找命令路径失败 %v", err)
		return err
	}
	logrus.Infof("找到命令路径 %s", path)
	// 替换当前进程镜像，执行用户传入的命令（不会返回）
	// syscall.Exec 是一个系统调用，用于执行指定的命令
	// cmdArray[0:] 是命令和参数的切片，os.Environ() 是当前进程的环境变量
	// cmdArray[0:]是安全写法，可以确保传入的是一个全新的切片对象（底层共用数据，但是独立的切片值）
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		logrus.Errorf("Exec command error %v", err.Error())
	}
	return nil
}

// readUserCommand 从管道中读取用户命令
func readUserCommand() []string {
	// 创建一个新的文件描述符，指向管道的读端， uintptr(3)是指index为3的文件描述符
	pipe := os.NewFile(uintptr(3), "pipe")
	//读取管道中的数据
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		logrus.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}
