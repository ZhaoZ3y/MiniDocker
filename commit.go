package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os/exec"
)

func commitContainer(imageName string) {
	mntURL := "/root/mnt" // 挂载点路径
	// 获取当前工作目录
	imageTar := "/root/" + imageName + ".tar" // 镜像文件路径
	// 压缩当前工作目录为 tar 文件
	fmt.Printf("%s", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		log.Errorf("压缩失败: %v", err)
		return
	}
}
