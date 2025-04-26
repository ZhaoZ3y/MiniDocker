package main

import (
	"MiniDocker/cgroup"
	"MiniDocker/cgroup/subsystems"
	"MiniDocker/container"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

// Run 启动容器的主函数
// tty: 是否启用终端交互（即 docker run -it 的效果）
// commandArray: 用户希望在容器中执行的命令及参数
// volume: 挂载的宿主机目录路径，格式为 /宿主机路径:/容器路径
func Run(tty bool, commandArray []string, volume string, res *subsystems.ResourceConfig) {
	// 创建容器父进程，并建立与子进程（init 进程）的通信管道
	parent, writePipe := container.NewParentProcess(tty, volume)
	if parent == nil {
		log.Error("父进程创建失败")
		return
	}

	// 启动父进程（实际是 fork 出子进程，执行自身程序并传参 "init"）
	if err := parent.Start(); err != nil {
		log.Error(err)
		return
	}

	//创建一个新的 cgroup 管理器，命名为 "MiniDocker-Cgroup"
	cgroupManager := cgroup.NewCgroupManager("MiniDocker-Cgroup")
	defer cgroupManager.Destroy() // 函数结束时自动销毁 cgroup，防止资源泄漏
	//设置资源限制，例如内存、CPU 等，res 是一个 ResourceConfig 类型的结构体
	if err := cgroupManager.Set(res); err != nil {
		log.Error("资源限制设置失败", err)
		return
	}
	//将容器进程加入该 cgroup 中进行限制管理
	if err := cgroupManager.Apply(parent.Process.Pid); err != nil {
		log.Error("加入进程失败", err)
		return
	}

	// 向 init 进程通过管道发送用户命令（init 进程会读取这个命令并执行）
	sendInitCommand(commandArray, writePipe)

	if tty {
		// 等待容器进程执行完毕，阻塞等待子进程退出
		parent.Wait()
		// 以下是清理工作：
		mntURL := "/root/mnt"                              // 容器挂载点路径
		rootURL := "/root"                                 // 容器根目录
		container.DeleteWorkSpace(rootURL, mntURL, volume) // 删除挂载的工作空间
	}

	// 容器运行结束，退出主进程
	os.Exit(0)
}

// sendInitCommand 将用户输入的命令通过管道写入给子进程（init 进程）
// writePipe: 是父进程传给子进程的文件描述符，用于通信
func sendInitCommand(commandArray []string, writePipe *os.File) {
	// 将命令数组转换为空格拼接的字符串，如 ["/bin/echo", "hello"] => "/bin/echo hello"
	command := strings.Join(commandArray, " ")
	log.Infof("用户传入的命令：%s", command)

	// 写入命令到管道（传给子进程的 fd3）
	writePipe.WriteString(command)

	// 关闭写端管道，表示写入完成
	writePipe.Close()
}
