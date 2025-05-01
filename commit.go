package main

import (
	"MiniDocker/container"
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"time"
)

func commitContainer(containerName string, imageName string) {
	mntURL := fmt.Sprintf(container.MntURL, containerName) // 挂载点路径
	imageTar := container.RootURL + "/" + imageName + ".tar"

	// 构造 tar 命令，排除 proc、sys、dev
	cmdArgs := []string{
		"-czf", imageTar,
		"--exclude=proc", "--exclude=sys", "--exclude=dev",
		"-C", mntURL, ".",
	}

	// 设置超时（例如 60 秒）
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "tar", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logrus.Infof("正在打包容器 %s 到镜像 %s", containerName, imageTar)

	if err := cmd.Run(); err != nil {
		// 检查是否是超时
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			logrus.Errorf("压缩命令超时")
		} else {
			logrus.Errorf("压缩失败: %v", err)
		}
		return
	}

	logrus.Infof("容器 %s 已成功提交为镜像 %s", containerName, imageTar)
}
