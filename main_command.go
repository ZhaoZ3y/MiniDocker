package main

import (
	"MiniDocker/cgroup/subsystems"
	"MiniDocker/container"
	"MiniDocker/network"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"os/exec"
	"strings"
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
		// -v 参数：用于挂载宿主机目录到容器内部
		&cli.StringFlag{
			Name:  "v",
			Usage: "挂载目录，例如: -v /host/path:/container/path",
		},
		// -e 参数：用于设置环境变量
		&cli.StringSliceFlag{
			Name:  "e",
			Usage: "设置环境变量，例如: -e VAR=value",
		},
		// -net 参数：用于设置容器的网络配置
		&cli.StringFlag{
			Name:  "net",
			Usage: "设置容器的网络配置，例如: --net bridge",
		},
		// -p 参数：用于设置端口映射
		&cli.StringSliceFlag{
			Name:  "p",
			Usage: "设置端口映射，例如: -p 8080:80",
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

		// 获取镜像名
		imageName := commandArray[0]    // 第一个参数是镜像名
		commandArray = commandArray[1:] // 剩余的参数是命令和参数

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
		// envSlice 是一个字符串切片，用于存储环境变量
		envSlice := ctx.StringSlice("e") // 获取环境变量参数
		// 执行容器创建与运行逻辑
		Run(createTty, commandArray, volume, resConf, containerName, imageName, envSlice)
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
		if ctx.NArg() < 2 {
			return fmt.Errorf("缺少容器名称参数和镜像名称参数")
		}
		containerName := ctx.Args().Get(0) // 获取容器名称
		imageName := ctx.Args().Get(1)     // 获取镜像名称
		commitContainer(containerName, imageName)
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
		// 这里检查环境变量，表示我们已经在容器内部了
		if os.Getenv(ENV_EXEC_PID) != "" {
			logrus.Infof("pid callback pid %d", os.Getpid())

			// 强制确保当前在容器根目录下
			if err := os.Chdir("/"); err != nil {
				logrus.Errorf("切换到根目录失败: %v", err)
			}

			// 记录当前工作目录
			cwd, err := os.Getwd()
			if err == nil {
				logrus.Infof("当前工作目录: %s", cwd)
			} else {
				logrus.Errorf("获取工作目录失败: %v", err)
			}

			// 获取要执行的命令
			cmdStr := os.Getenv(ENV_EXEC_CMD)
			if cmdStr == "" {
				return fmt.Errorf("没有指定要执行的命令")
			}

			// 分割命令字符串为命令和参数
			// 这里需要特别注意空格分隔的处理
			cmdParts := strings.Fields(cmdStr) // 使用 Fields 更好地处理多个空格
			if len(cmdParts) == 0 {
				return fmt.Errorf("命令为空")
			}

			// 创建命令
			cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Dir = "/" // 强制设置工作目录为根目录

			// 执行命令
			return cmd.Run()
		}

		// 原有的代码，处理非容器内部的情况
		if ctx.NArg() < 2 {
			return fmt.Errorf("缺少容器名称和命令参数")
		}
		containerName := ctx.Args().Get(0)
		var commandArray []string
		for _, arg := range ctx.Args().Tail() {
			commandArray = append(commandArray, arg)
		}

		// 调用 ExecContainer 函数
		ExecContainer(containerName, commandArray)
		return nil
	},
}

// stopCommand 命令定义：停止容器
var stopCommand = &cli.Command{
	Name:  "stop",
	Usage: "停止容器",
	Action: func(ctx *cli.Context) error {
		// 参数检查：至少需要一个容器名称参数
		if ctx.NArg() < 1 {
			logrus.Error("缺少容器名称参数")
		}
		containerName := ctx.Args().Get(0) // 获取容器名称
		// 停止容器
		stopContainer(containerName)
		return nil
	},
}

// removeCommand 命令定义：删除容器
var removeCommand = &cli.Command{
	Name:  "rm",
	Usage: "删除容器",
	Action: func(ctx *cli.Context) error {
		// 参数检查：至少需要一个容器名称参数
		if ctx.NArg() < 1 {
			logrus.Error("缺少容器名称参数")
		}
		containerName := ctx.Args().Get(0) // 获取容器名称
		// 删除容器
		removeContainer(containerName)
		return nil
	},
}

// networkCommand 命令定义：网络相关命令
var networkCommand = &cli.Command{
	Name:  "network",
	Usage: "网络相关命令",
	Subcommands: []*cli.Command{
		{
			Name:  "create",
			Usage: "创建网络",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "driver",
					Usage: "网络驱动",
				},
				&cli.StringFlag{
					Name:  "subnet",
					Usage: "子网地址",
				},
			},
			Action: func(ctx *cli.Context) error {
				// 参数检查：至少需要一个网络名称参数
				if ctx.NArg() < 1 {
					return fmt.Errorf("缺少网络名称参数")
				}
				network.Init() // 初始化网络
				err := network.CreateNetWork(ctx.String("driver"), ctx.String("subnet"), ctx.Args().Get(0))
				if err != nil {
					return fmt.Errorf("创建网络失败: %v", err)
				}
				return nil
			},
		},
		{
			Name:  "list",
			Usage: "列出网络",
			Action: func(ctx *cli.Context) error {
				network.Init()        // 初始化网络
				network.ListNetwork() // 列出网络
				return nil
			},
		},
		{
			Name:  "remove",
			Usage: "删除网络",
			Action: func(ctx *cli.Context) error {
				// 参数检查：至少需要一个网络名称参数
				if ctx.NArg() < 1 {
					return fmt.Errorf("缺少网络名称参数")
				}
				network.Init() // 初始化网络
				err := network.DeleteNetwork(ctx.Args().Get(0))
				if err != nil {
					return fmt.Errorf("删除网络失败: %v", err)
				}
				return nil
			},
		},
	},
}
