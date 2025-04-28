package main

import (
	"MiniDocker/cgroup/subsystems"
	"MiniDocker/container"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
)

// runCommand 命令定义：用于创建并运行一个容器
// 使用示例：MiniDocker run -ti /bin/bash
var runCommand = &cli.Command{
	Name:  "run", // 命令名称为 run
	Usage: `创建一个容器，并启用 namespace 和 cgroups 资源限制，例如: MiniDocker run -ti [镜像名] [命令]`,
	Flags: []cli.Flag{
		// -ti 参数：表示是否启用 tty 和交互模式
		&cli.BoolFlag{
			Name:  "ti",
			Usage: "启用 tty 和交互模式（interactive mode），即类似 docker run -it",
		},
		// -v 参数：用于挂载宿主机目录到容器内部
		&cli.StringFlag{
			Name:  "v",
			Usage: "挂载目录，例如: -v /host/path:/container/path",
		},
		// -d 参数：表示是否在后台运行容器
		&cli.BoolFlag{
			Name:  "d",
			Usage: "后台运行容器",
		},
		// -m 参数：用于设置容器的内存限制
		&cli.StringFlag{
			Name:  "m",
			Usage: "设置容器的内存限制，例如: -m 512m",
		},
		// -cpushare 参数：用于设置容器的 CPU 限制
		&cli.StringFlag{
			Name:  "cpushare",
			Usage: "设置容器的 CPU 限制，例如: -cpushare 512",
		},
		// -cpuset 参数：用于设置容器的 CPU 核心限制
		&cli.StringFlag{
			Name:  "cpuset",
			Usage: "设置容器的 CPU 核心限制，例如: -cpuset 0,1",
		},
		// --name 参数：用于设置容器的名称
		&cli.StringFlag{
			Name:  "name",
			Usage: "设置容器的名称，例如: --name my_container",
		},
	},
	Action: func(ctx *cli.Context) error {
		// 参数检查：至少需要一个命令参数（即用户要运行的程序）
		if ctx.NArg() < 1 {
			return fmt.Errorf("缺少要执行的命令参数")
		}
		// 获取完整的命令数组（包含命令和其参数）
		var commandArray []string
		for _, arg := range ctx.Args().Slice() {
			commandArray = append(commandArray, arg)
		}

		// 获取是否启用 tty 和交互模式（布尔值）
		createTty := ctx.Bool("ti")
		detach := ctx.Bool("d") // 是否后台运行容器
		// createTty 和 detach 互斥，不能同时为 true
		if createTty && detach {
			return fmt.Errorf("不能同时使用 -ti 和 -d 参数")
		}
		// resConf 是资源限制配置结构体，包含内存、CPU 权重和 CPU 核心限制
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: ctx.String("m"),        // 内存限制
			CpuShare:    ctx.String("cpushare"), // CPU 权重限制
			CpuSet:      ctx.String("cpuset"),   // CPU 核心限制
		}
		logrus.Infof("createTty: %v", createTty)
		// 获取容器名称参数
		containerName := ctx.String("name")
		// 获取用户指定的挂载目录（形如 /宿主机路径:/容器路径）
		volume := ctx.String("v")

		// 执行容器创建与运行逻辑
		Run(createTty, commandArray, volume, resConf, containerName)
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
		logrus.Infof("初始化容器")
		// 调用容器的初始化进程
		err := container.RunContainerInitProcess()
		return err
	},
}

// commitCommand 命令定义：提交容器的更改为新的镜像
var commitCommand = &cli.Command{
	Name:  "commit",
	Usage: "提交容器的更改",
	Action: func(ctx *cli.Context) error {
		if ctx.NArg() < 1 {
			return fmt.Errorf("缺少容器名称参数")
		}
		imageName := ctx.Args().Get(0) // 获取容器名称
		commitContainer(imageName)
		return nil
	},
}

// listCommand 命令定义：列出所有容器
var listCommand = &cli.Command{
	Name:  "ps",
	Usage: "列出所有容器",
	Action: func(ctx *cli.Context) error {
		ListContainers()
		return nil
	},
}

// logCommand 命令定义：查看容器的日志
var logCommand = &cli.Command{
	Name:  "logs",
	Usage: "查看容器的日志",
	Action: func(ctx *cli.Context) error {
		if ctx.NArg() < 1 {
			return fmt.Errorf("缺少容器名称参数")
		}
		containerName := ctx.Args().Get(0) // 获取容器名称
		logContainer(containerName)
		return nil
	},
}

// execCommand 命令定义：在容器中执行命令
var execCommand = &cli.Command{
	Name:  "exec",
	Usage: "在容器中执行命令",
	Action: func(ctx *cli.Context) error {
		if os.Getenv(ENV_EXEC_PID) != "" {
			logrus.Infof("pid callback pid %s", os.Getgid())
			return nil
		}
		if ctx.NArg() < 2 {
			return fmt.Errorf("缺少容器名称和命令参数")
		}
		containerName := ctx.Args().Get(0) // 获取容器名称
		var commandArray []string
		// 将除了第一个参数（容器名称）之外的所有参数都加入命令数组
		for _, arg := range ctx.Args().Tail() {
			commandArray = append(commandArray, arg)
		}
		// 执行容器中的命令
		ExecContainer(containerName, commandArray)
		return nil
	},
}
