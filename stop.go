package main

import (
	"MiniDocker/container"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"
)

// stopContainer 函数：停止指定名称的容器
func stopContainer(containerName string) {
	// 根据容器名称获取pid
	pid, err := GetContainerPidByName(containerName)
	if err != nil {
		logrus.Errorf("获取容器 %s 的 PID 失败: %v", containerName, err)
		return
	}
	// 将 pid 转换为整数
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		logrus.Errorf("PID 转换失败: %v", err)
		return
	}

	// 检查 PID 是否存在
	_, err = os.FindProcess(pidInt)
	if err != nil {
		logrus.Errorf("无法找到容器 %s 的进程: %v", containerName, err)
		return
	}

	logrus.Infof("容器 %s 进程存在，准备发送信号", containerName)

	// 发送 SIGTERM 信号给容器进程，优雅停止
	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		logrus.Errorf("停止容器 %s 失败: %v", containerName, err)
		return
	}

	// 跟据容器名称获取容器信息
	info, err := getContainerInfoByName(containerName)
	if err != nil {
		logrus.Errorf("获取容器信息失败: %v", err)
		return
	}
	// 容器进程已经停止，修改容器状态和PID
	info.Status = container.STOPPED
	info.Pid = ""
	// 更新容器信息
	newContentBytes, err := json.Marshal(info)
	if err != nil {
		logrus.Errorf("容器信息序列化失败: %v", err)
		return
	}
	// 获取容器信息的目录路径
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName
	// 写入新的容器信息
	if err := ioutil.WriteFile(configFilePath, newContentBytes, 0622); err != nil {
		logrus.Errorf("写入容器信息失败: %v", err)
		return
	}
	logrus.Infof("容器 %s 停止成功", containerName)
}
