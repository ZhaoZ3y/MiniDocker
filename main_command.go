package main

import (
	"MiniDocker/container"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// runCommand 命令定义：用于创建并运行一个容器
// 使用示例：MiniDocker run -ti /bin/bash
var runCommand = &cli.Command{
	Name:  "run",
	Usage: `创建一个容器，并启用 namespace 和 cgroups 资源限制，例如: MiniDocker run -ti [镜像名] [命令]`,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "ti",
			Usage: "启用 tty 和交互模式（interactive mode），即类似 docker run -it",
		},
	},
	Action: func(ctx *cli.Context) error {
		// 参数检查：至少需要一个命令参数
		if ctx.NArg() < 1 {
			return fmt.Errorf("缺少容器名称参数")
		}
		// 获取执行的命令
		cmd := ctx.Args().Get(0)
		// 获取是否启用 tty 和交互模式
		tty := ctx.Bool("ti")
		Run(tty, cmd)
		return nil
	},
}

// initCommand 命令定义：容器内部使用的初始化命令
// 注意：这个命令不是用户手动调用的，而是由父进程在容器环境中自动触发
var initCommand = &cli.Command{
	Name:  "init",
	Usage: `初始化容器，创建 namespace 和 cgroups 资源限制，例如: MiniDocker init [容器名称]`,
	Action: func(ctx *cli.Context) error {
		// 日志记录：进入容器初始化流程
		log.Infof("初始化容器")
		// 获取容器初始化命令（用户传入的要执行的命令）
		cmd := ctx.Args().Get(0)
		log.Infof("command %s", cmd)
		// 调用容器的初始化进程
		err := container.RunContainerInitProcess(cmd, nil)
		return err
	},
}
