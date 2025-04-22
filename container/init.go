package container

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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

	// 设置挂载点
	setUpMount()
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

// pivotRoot 执行 pivot_root 系统调用，切换当前进程的根文件系统
func pivotRoot(root string) error {
	// 1. 先把 root 重新 mount 一次，把自己挂载到自己上，目的是为了创建一个新的 mount namespace，
	// 避免出现 “device or resource busy” 的问题
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		logrus.Errorf("设置 mount namespace 私有失败: %v", err)
	}

	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		logrus.Errorf("重新 bind mount root 失败: %v", err)
	}

	// 2. 创建一个 .pivot_root 目录用于存放旧的 root，
	// pivot_root 系统调用要求第二个参数必须在新的 root 目录内部
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0755); err != nil {
		return fmt.Errorf("创建 pivot_root 目录失败: %v", err)
	}

	// 3. 执行 pivot_root，实际上会把当前 root 移到 .pivot_root，
	// 并将新的 root 设置为 root（即 busybox 解压目录）
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("执行 pivot_root 失败: %v", err)
	}

	// 4. 切换当前进程的工作目录到新的根目录，否则后续某些相对路径操作会出问题
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("切换工作目录到新 root 失败: %v", err)
	}

	// 5. 重新定义 pivotDir，这时候它已经在新的 root 环境下，
	// 需要用新的路径表示它
	pivotDir = filepath.Join("/", ".pivot_root")

	// 6. 卸载旧的 root（现在被挂载在 /.pivot_root），使用 MNT_DETACH 表示延迟卸载，直到没有进程使用
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("卸载旧的 root 失败: %v", err)
	}

	// 7. 删除 .pivot_root 目录，清理现场
	return os.Remove(pivotDir)
}

// setUpMount 设置容器的挂载点
// 主要是挂载 /proc 和 /dev 目录
func setUpMount() {
	// 获取当前工作目录
	pwd, err := os.Getwd()
	if err != nil {
		logrus.Errorf("获取当前工作目录失败: %v", err)
		return
	}
	logrus.Infof("当前工作目录: %s", pwd)
	// 将当前目录重新挂载一次，作为挂载点
	if err := syscall.Mount(pwd, pwd, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		logrus.Errorf("重新绑定挂载点失败: %v", err)
		return
	}

	// 执行 pivot_root 切换根文件系统
	if err := pivotRoot(pwd); err != nil {
		logrus.Errorf("执行 pivot_root 失败: %v", err)
		return
	}
	// 设置默认挂载参数：
	// MS_NOEXEC：不允许执行二进制
	// MS_NOSUID：不允许 set-user-ID 或 set-group-ID
	// MS_NODEV：不允许访问设备文件
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV

	// 挂载 proc 文件系统到 /proc 目录
	// /proc 提供了内核和进程相关的虚拟信息，容器中 ps/top 等命令依赖它
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		logrus.Errorf("挂载 /proc 失败: %v", err)
		return
	}
	// 挂载 tmpfs 到 /dev 目录
	if err := syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755"); err != nil {
		logrus.Errorf("挂载 /dev 失败: %v", err)
		return
	}
}
