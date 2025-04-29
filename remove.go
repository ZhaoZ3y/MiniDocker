package main

import (
	"MiniDocker/container"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

func removeContainer(containerName string) {
	// 根据容器名称获取容器信息
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		logrus.Errorf("获取容器信息失败: %v", err)
		return
	}
	// 只删除处于停止状态的容器
	if containerInfo.Status != container.STOPPED {
		logrus.Errorf("容器 %s 处于运行状态，无法删除", containerName)
		return
	}
	// 删除存储容器信息的目录
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.RemoveAll(dirURL); err != nil {
		logrus.Errorf("删除容器目录失败: %v", err)
		return
	}
}
