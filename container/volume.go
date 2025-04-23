package container

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

// NewWorkSpace 创建容器工作空间：包括只读层、写层和挂载点
func NewWorkSpace(rootURL string, mountURL string) {
	CreateReadOnlyLayer(rootURL)        // 创建只读层（busybox）
	CreateWriteLayer(rootURL)           // 创建写层目录
	CreateMountPoint(rootURL, mountURL) // 将只读层和写层通过 aufs 挂载到挂载点
}

// CreateReadOnlyLayer 创建只读层（解压 busybox.tar）
func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := rootURL + "busybox/"       // 解压路径
	busyboxTarURL := rootURL + "busybox.tar" // busybox 压缩包路径

	// 判断只读层目录是否存在
	exist, err := PathExists(busyboxURL)
	if err != nil {
		log.Errorf("判断目录 %s 是否存在失败: %v", busyboxURL, err)
	}

	// 如果不存在则创建目录并解压 busybox
	if exist == false {
		if err := os.Mkdir(busyboxURL, 0777); err != nil {
			log.Errorf("创建目录 %s 失败: %v", busyboxURL, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			log.Errorf("解压目录 %s 失败: %v", busyboxURL, err)
		}
	}
}

// CreateWriteLayer 创建写层目录
func CreateWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/" // 写层目录路径
	if err := os.Mkdir(writeURL, 0777); err != nil {
		log.Errorf("创建写层目录 %s 失败: %v", writeURL, err)
	}
}

// CreateMountPoint 使用 aufs 将只读层和写层挂载为联合文件系统
func CreateMountPoint(rootURL string, mountURL string) {
	// 创建挂载点目录
	if err := os.Mkdir(mountURL, 0777); err != nil {
		log.Errorf("创建挂载点目录 %s 失败: %v", mountURL, err)
	}

	// 构造 aufs 挂载参数：写层在前，只读层在后
	dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"

	// 执行 mount 命令挂载 aufs 文件系统
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mountURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("挂载失败: %v", err)
	}
}

// PathExists 判断路径是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil // 路径存在
	}
	if os.IsNotExist(err) {
		return false, nil // 路径不存在
	}
	return false, err // 其他错误
}

// DeleteWorkSpace 删除容器工作空间：包括卸载挂载点和删除写层
func DeleteWorkSpace(rootURL string, mountURL string) {
	DeleteMountPoint(rootURL, mountURL) // 卸载挂载点并删除挂载目录
	DeleteWriteLayer(rootURL)           // 删除写层目录
}

// DeleteMountPoint 卸载 aufs 文件系统并删除挂载目录
func DeleteMountPoint(rootURL string, mountURL string) {
	// 执行 umount 命令卸载挂载点
	cmd := exec.Command("umount", mountURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("卸载挂载点 %s 失败: %v", mountURL, err)
	}

	// 删除挂载点目录
	if err := os.RemoveAll(mountURL); err != nil {
		log.Errorf("删除挂载点目录 %s 失败: %v", mountURL, err)
	}
}

// DeleteWriteLayer 删除写层目录
func DeleteWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("删除写层目录 %s 失败: %v", writeURL, err)
	}
}
