package main

import (
	"MiniDocker/container"
	log "github.com/sirupsen/logrus"
	"os"
)

// Run 开启一个新的父进程（容器进程），并在其中执行指定的命令
// tty 参数表示是否启用终端交互模式（对应 docker run -it）
// command 是用户想要在容器中执行的命令（如 /bin/bash）
func Run(tty bool, command string) {
	// 创建一个新的父进程，实际上是再次调用自身程序执行 "init" 子命令
	// 在 container.NewParentProcess 中设置了 namespace 和 tty 等
	parent := container.NewParentProcess(tty, command)
	if err := parent.Start(); err != nil {
		log.Error(err)
	}
	parent.Wait()
	os.Exit(-1)
}
