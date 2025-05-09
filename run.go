package main

import (
	"MiniDocker/cgroup"
	"MiniDocker/cgroup/subsystems"
	"MiniDocker/container"
	"MiniDocker/network"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// Run 启动一个容器实例
// tty 表示是否绑定终端（类似 docker run -it）
// commandArray 是用户希望在容器中执行的命令及参数
// volume 是宿主机与容器的挂载路径
func Run(tty bool, commandArray []string, volume string, res *subsystems.ResourceConfig, containerName string, imageName string, envSlice []string, nw string, portmapping []string) {
	containerID := randStringBytes(10)
	if containerName == "" {
		containerName = containerID
	}
	// 创建容器父进程和通信管道
	parent, writePipe := container.NewParentProcess(tty, volume, containerName, imageName, envSlice)
	if parent == nil {
		logrus.Error("父进程创建失败")
		return
	}
	// 启动父进程（fork 自身，进入 init 子流程）
	if err := parent.Start(); err != nil {
		logrus.Error(err)
		return
	}

	// 记录容器基本信息
	containerName, err := recordContainerInfo(parent.Process.Pid, commandArray, containerName, containerID, volume)
	if err != nil {
		logrus.Error("容器信息记录失败", err)
		return
	}

	// 创建并配置 Cgroup 管理器
	cgroupManager := cgroup.NewCgroupManager("MiniDocker-Cgroup")
	defer cgroupManager.Destroy() // 确保退出时清理资源
	// 设置 Cgroup 资源限制
	cgroupManager.Set(res)
	// 将容器进程加入 Cgroup
	cgroupManager.Apply(parent.Process.Pid)

	if nw != "" {
		// 初始化网络配置
		network.Init()
		// 配置容器网络
		containerInfo := &container.Info{
			Id:          containerID,
			Pid:         strconv.Itoa(parent.Process.Pid),
			Name:        containerName,
			PortMapping: portmapping,
		}
		if err := network.Connect(nw, containerInfo); err != nil {
			logrus.Error("网络连接失败", err)
			return
		}
	}

	// 发送用户命令给 init 进程执行
	sendInitCommand(commandArray, writePipe)

	if tty {
		// 前台模式，等待容器退出
		parent.Wait()
		deleteContainerInfo(containerName)
		container.DeleteWorkSpace(volume, containerName) // 删除容器工作空间
	}
}

// sendInitCommand 将用户命令写入管道，传递给子进程（init 进程）
func sendInitCommand(commandArray []string, writePipe *os.File) {
	command := strings.Join(commandArray, " ")
	logrus.Infof("用户传入的命令：%s", command)

	writePipe.WriteString(command)
	writePipe.Close()
}

// recordContainerInfo 保存容器信息到本地
func recordContainerInfo(containerPID int, commandAry []string, containerName string, id string, volume string) (string, error) {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandAry, "")

	containerInfo := &container.Info{
		Id:          id,
		Pid:         strconv.Itoa(containerPID),
		Command:     command,
		CreatedTime: currentTime,
		Status:      container.RUNNING,
		Name:        containerName,
		Volume:      volume,
	}

	containerInfoJson, err := json.Marshal(containerInfo)
	if err != nil {
		logrus.Errorf("容器信息序列化失败: %v", err)
		return "", err
	}

	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.MkdirAll(dirURL, 0622); err != nil {
		logrus.Errorf("创建目录失败: %v", err)
		return "", err
	}

	fileName := path.Join(dirURL, container.ConfigName)
	file, err := os.Create(fileName)
	if err != nil {
		logrus.Errorf("创建文件失败: %v", err)
		return "", err
	}

	if _, err := file.WriteString(string(containerInfoJson)); err != nil {
		logrus.Errorf("写入容器信息失败: %v", err)
		return "", err
	}

	return containerName, nil
}

// randStringBytes 生成指定长度的随机字符串（仅包含数字）
func randStringBytes(n int) string {
	letteBytes := []byte("0123456789")
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letteBytes[rand.Intn(len(letteBytes))]
	}
	return string(b)
}

// deleteContainerInfo 删除本地保存的容器信息
func deleteContainerInfo(containerId string) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirURL); err != nil {
		logrus.Errorf("删除容器信息目录失败: %v", err)
	}
}
