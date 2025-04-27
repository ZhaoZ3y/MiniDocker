package main

import (
	"MiniDocker/container"
	"encoding/json"
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

// getContainerInfo 读取指定容器的配置文件，解析并返回容器信息
func getContainerInfo(file os.FileInfo) (*container.Info, error) {
	// 获取文件名
	containerName := file.Name()
	// 拼接完整路径
	configFilePath := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath = configFilePath + container.ConfigName
	// 读取配置文件内容
	content, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		logrus.Errorf("读取配置文件 %v 失败: %v", configFilePath, err)
		return nil, err
	}
	// 解析 JSON 内容为 Info 结构体
	var containerInfo container.Info
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		logrus.Errorf("解析 JSON 失败: %v", err)
		return nil, err
	}
	// 返回解析后的容器信息
	return &containerInfo, nil
}
