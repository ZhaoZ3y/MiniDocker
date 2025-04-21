package main

import (
	"MiniDocker/container"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

// Run 开启一个新的父进程（容器进程），并在其中执行指定的命令
// tty 参数表示是否启用终端交互模式（对应 docker run -it）
// commandArray 是用户想要在容器中执行的命令（如 ["/bin/bash"]）
// res 是 cgroup 的资源配置，用于限制容器的资源使用
func Run(tty bool, commandArray []string) {
	// 创建一个新的父进程，实际上是再次调用自身程序执行 "init" 子命令
	// 创建一个新的父进程，返回 exec.Cmd 实例和用于通信的管道写端
	parent, writePipe := container.NewParentProcess(tty)
	if parent == nil {
		log.Errorf("新建父进程失败")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	//// 创建cgroup Manager 实例
	//cgroupManager := Cgroups.NewCgroupManager("MiniDocker-cgroup")
	//defer cgroupManager.Destroy()
	//// 设置资源限制
	//if err := cgroupManager.Set(res); err != nil {
	//	log.Errorf("设置资源限制失败 %v", err)
	//	return
	//}
	//// 将当前进程加入到 cgroup 中
	//if err := cgroupManager.Apply(parent.Process.Pid); err != nil {
	//	log.Errorf("将进程加入 cgroup 失败 %v", err)
	//	return
	//}
	//
	// 初始化容器进程
	sendInitCommand(commandArray, writePipe)
	parent.Wait()
	os.Exit(0)
}

// sendInitCommand 向容器的管道写入初始化命令
// 这个命令会在容器中执行，通常是一个 shell 命令（如 /bin/bash）
// 它会在容器的命名空间中执行，并且可以通过 tty 进行交互
func sendInitCommand(commandArray []string, writePipe *os.File) {
	// 将命令数组转换为字符串，以空格分隔
	cmdStr := strings.Join(commandArray, " ")
	// 将命令写入管道
	if _, err := writePipe.WriteString(cmdStr); err != nil {
		log.Errorf("写入命令到管道失败 %v", err)
	}
	writePipe.Close()
}
