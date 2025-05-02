package main

import (
	"MiniDocker/container"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
)

// GetContainerPidByName 根据容器名字获取其 PID
func GetContainerPidByName(containerName string) (string, error) {
	// 拼接出容器对应的配置文件路径
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName

	// 读取容器的配置信息
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}

	var containerInfo container.Info
	// 反序列化 JSON 数据到 containerInfo 结构体
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return "", err
	}

	// 返回容器的 PID
	return containerInfo.Pid, nil
}

// getContainerInfoByName 函数：根据容器名称获取容器信息
func getContainerInfoByName(containerName string) (*container.Info, error) {
	// 获取容器信息的目录路径
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName
	// 读取容器信息
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		logrus.Errorf("读取容器信息失败: %v", err)
		return nil, err
	}
	// 解析容器信息
	var info container.Info
	if err := json.Unmarshal(contentBytes, &info); err != nil {
		logrus.Errorf("解析容器信息失败: %v", err)
		return nil, err
	}
	// 返回容器信息
	return &info, nil
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

// getEnvsByPid 读取指定 PID 的环境变量
func getEnvsByPid(pid string) ([]string, error) {
	// 进程环境变量的路径
	path := fmt.Sprintf("/proc/%s/environ", pid)
	// 读取环境变量文件
	contentBytes, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.Errorf("读取环境变量失败: %v", err)
		return nil, err
	}
	// 分割环境变量字符串
	envs := strings.Split(string(contentBytes), "\u0000")
	// 返回环境变量列表
	return envs, nil
}
