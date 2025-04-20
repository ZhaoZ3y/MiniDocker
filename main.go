package main

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

const usage = "MiniDocker是一个简单用Go实现的仿Docker。\n" +
	"目的是为了使用这个项目学习一下docker的原理以及拿来当简历上的轮子项目。"

// 函数启动入口
func main() {
	// 初始化命令行应用创建实例
	app := &cli.App{
		Name:  "MiniDocker",
		Usage: usage,
		// 主要的命令
		Commands: []*cli.Command{
			initCommand,
			runCommand,
		},
		// 执行命令前的日志输出
		Before: func(c *cli.Context) error {
			logrus.SetFormatter(&logrus.TextFormatter{})
			logrus.SetOutput(os.Stdout)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
