package main

import (
	"MiniDocker/container"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"text/tabwriter"
)

// ListContainers 列出当前所有容器及其状态
func ListContainers() {
	// 容器信息根目录
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, "")
	dirURL = dirURL[:len(dirURL)-1]

	// 读取该目录下的所有子目录（每个容器一个目录）
	files, err := ioutil.ReadDir(dirURL)
	if err != nil {
		logrus.Errorf("读取目录失败: %v", err)
		return
	}

	var containers []*container.Info
	// 遍历每个子目录，解析容器信息
	for _, file := range files {
		// 根据配置文件名获取容器信息
		tmpContainer, err := getContainerInfo(file)
		if err != nil {
			logrus.Errorf("获取容器信息失败: %v", err)
			continue
		}
		// 解析容器信息
		containers = append(containers, tmpContainer)
	}
	// 使用表格格式打印容器信息
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATE\n")
	for _, c := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			c.Id,
			c.Name,
			c.Pid,
			c.Status,
			c.Command,
			c.CreatedTime,
		)
	}
	if err := w.Flush(); err != nil {
		logrus.Errorf("刷新表格失败: %v", err)
		return
	}
}
