package main

import (
	"MiniDocker/container"
	log "github.com/sirupsen/logrus"
	"os"
)

// Run 开启一个新的父进程（容器进程），并在其中执行指定的命令
// 参数 tty 表示是否启用终端交互模式（即 docker run -it）
// 参数 command 是用户想要在容器中执行的命令（如 /bin/bash）
func Run(tty bool, command string) {
	// 创建一个新的父进程（容器进程）
	// 实质上是调用 /proc/self/exe 再次执行自己，但附带参数 "init"，进入初始化流程
	// 在 container.NewParentProcess 中会设置 namespace、是否绑定 tty、初始化命令等
	parent := container.NewParentProcess(tty, command)

	// 启动父进程（此时会运行新的子进程，也就是容器进程）
	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	// 等待子进程退出（也就是容器进程结束）
	parent.Wait()

	// 主进程退出，容器运行结束
	os.Exit(-1)
}
