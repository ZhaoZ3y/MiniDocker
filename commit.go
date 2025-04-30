package main

import (
	"MiniDocker/container"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os/exec"
)

func commitContainer(containerName string, imageName string) {
	mntURL := fmt.Sprintf(container.MntURL, containerName) // 挂载点路径
	mntURL += "/"                                          // 在路径后添加斜杠，以确保 tar 命令正确处理目录
	// 获取当前工作目录
	imageTar := container.RootURL + "/" + imageName + ".tar" // 镜像文件路径
	// 压缩当前工作目录为 tar 文件
	fmt.Printf("%s", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		log.Errorf("压缩失败: %v", err)
		return
	}
}
