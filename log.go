package main

import (
	"MiniDocker/container"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

// logContainer 查看指定容器的日志
func logContainer(containerName string) {
	// 拼接容器日志文件路径
	logFilePath := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	logFileLocation := logFilePath + container.ContainerLogFile
	// 打开日志文件
	file, err := os.Open(logFileLocation)
	defer file.Close()
	if err != nil {
		logrus.Errorf("打开日志文件失败: %v", err)
		return
	}
	// 读取文件内容
	content, err := ioutil.ReadAll(file)
	if err != nil {
		logrus.Errorf("读取日志文件失败: %v", err)
		return
	}
	// 打印日志内容
	fmt.Fprint(os.Stdout, string(content))
}
