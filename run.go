package main

import (
	"MiniDocker/container"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

// Run 启动容器的主函数
// tty: 是否启用终端交互（即 docker run -it 的效果）
// commandArray: 用户希望在容器中执行的命令及参数
func Run(tty bool, commandArray []string) {
	// 创建容器父进程，并建立与子进程的通信管道
	parent, writePipe := container.NewParentProcess(tty)
	if parent == nil {
		log.Error("父进程创建失败")
		return
	}

	// 启动容器进程（即再次执行自身程序，参数为 "init"）
	if err := parent.Start(); err != nil {
		log.Error(err)
		return
	}

	// ========== 以下是用于设置 cgroup 的资源限制逻辑，暂时注释掉 ==========
	//cgroupManager := cgroup.NewCgroupManager("MiniDocker-Cgroup")
	//defer cgroupManager.Destroy() // 进程退出时自动销毁 cgroup
	//
	//// 设置资源限制（如内存、CPU等），res 是一个 ResourceConfig 配置结构体
	//if err := cgroupManager.Set(res); err != nil {
	//	log.Error("资源限制设置失败", err)
	//	return
	//}
	//
	//// 将容器进程加入 cgroup 管理
	//if err := cgroupManager.Apply(parent.Process.Pid); err != nil {
	//	log.Error("加入进程失败", err)
	//	return
	//}
	// =======================================================

	// 将用户的命令写入到管道中，传递给 init 进程
	sendInitCommand(commandArray, writePipe)

	// 等待容器进程结束（父进程阻塞）
	parent.Wait()

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
