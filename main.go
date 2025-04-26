package main

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

// usage 是 CLI 应用的简介，用于帮助信息中显示
const usage = "MiniDocker是一个简单用Go实现的仿Docker。\n" +
	"目的是为了使用这个项目学习一下docker的原理以及拿来当简历上的轮子项目。"

// main 是程序的入口函数
func main() {
	// 创建 CLI 应用实例
	app := &cli.App{
		Name:  "MiniDocker", // 应用名称
		Usage: usage,        // 应用说明
		Commands: []*cli.Command{
			initCommand,   // 初始化容器（由容器进程自动调用）
			runCommand,    // 创建并运行容器（用户调用）
			commitCommand, // 提交容器（用户调用）
			listCommand,   // 列出容器（用户调用）
		},
		// 在执行命令前统一设置日志格式和输出目标
		Before: func(c *cli.Context) error {
			logrus.SetFormatter(&logrus.TextFormatter{}) // 设置为文本日志格式
			logrus.SetOutput(os.Stdout)                  // 输出日志到标准输出
			return nil
		},
	}

	// 启动应用，解析 os.Args 中的命令和参数
	if err := app.Run(os.Args); err != nil {
		// 若发生错误，记录日志并退出程序
		log.Fatal(err)
	}
}
