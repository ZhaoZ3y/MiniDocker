package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

// NewWorkSpace 创建容器工作空间，包括只读层（busybox）、写层和挂载点。
// rootURL 是工作空间的根路径，mountURL 是容器挂载点路径。
func NewWorkSpace(rootURL string, mountURL string) {
	CreateReadOnlyLayer(rootURL)        // 解压 busybox 创建只读层
	CreateWriteLayer(rootURL)           // 创建写层目录（upper、work）
	CreateMountPoint(rootURL, mountURL) // 使用 OverlayFS 挂载为联合文件系统

	containerRootDir := mountURL + "/root"
	if err := os.MkdirAll(containerRootDir, 0755); err != nil {
		log.Errorf("创建容器内 /root 目录失败: %v", err)
	}
}

// CreateReadOnlyLayer 创建只读层：解压 busybox.tar 到指定目录。
func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := rootURL + "busybox/"       // 解压路径
	busyboxTarURL := rootURL + "busybox.tar" // busybox 压缩包路径

	exist, err := PathExists(busyboxURL)
	if err != nil {
		log.Errorf("判断目录 %s 是否存在失败: %v", busyboxURL, err)
	}

	// 如果 busybox 目录不存在，创建目录并解压
	if !exist {
		if err := os.Mkdir(busyboxURL, 0777); err != nil {
			log.Errorf("创建目录 %s 失败: %v", busyboxURL, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			log.Errorf("解压目录 %s 失败: %v", busyboxURL, err)
		}
	}
}

// CreateWriteLayer 创建写层目录，包括 overlay 所需的 upper 和 work 子目录。
func CreateWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.MkdirAll(writeURL, 0777); err != nil && !os.IsExist(err) {
		log.Errorf("创建写层目录 %s 失败: %v", writeURL, err)
	}
}

// CreateMountPoint 使用 OverlayFS 将只读层和写层挂载到挂载点目录。
func CreateMountPoint(rootURL string, mountURL string) {
	if err := os.MkdirAll(mountURL, 0777); err != nil && !os.IsExist(err) {
		log.Errorf("创建挂载点目录 %s 失败: %v", mountURL, err)
	}

	lowerDir := rootURL + "busybox"          // 只读层目录
	upperDir := rootURL + "writeLayer/upper" // 写层 upper 目录
	workDir := rootURL + "writeLayer/work"   // OverlayFS 工作目录

	// 创建 upper 和 work 子目录
	if err := os.MkdirAll(upperDir, 0777); err != nil {
		log.Errorf("创建 upper 目录失败: %v", err)
	}
	if err := os.MkdirAll(workDir, 0777); err != nil {
		log.Errorf("创建 work 目录失败: %v", err)
	}

	// 构造 overlay 挂载参数，并执行 mount 命令
	options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerDir, upperDir, workDir)
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", options, mountURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("挂载 OverlayFS 失败: %v", err)
	}
}

// PathExists 判断指定路径是否存在。
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

// DeleteWorkSpace 删除容器工作空间，卸载挂载点并清理写层目录。
func DeleteWorkSpace(rootURL string, mountURL string) {
	DeleteMountPoint(rootURL, mountURL) // 卸载挂载点
	DeleteWriteLayer(rootURL)           // 删除写层
}

// DeleteMountPoint 卸载挂载点，并删除挂载点目录。
func DeleteMountPoint(rootURL string, mountURL string) {
	cmd := exec.Command("umount", mountURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("卸载挂载点 %s 失败: %v", mountURL, err)
	}

	if err := os.RemoveAll(mountURL); err != nil {
		log.Errorf("删除挂载点目录 %s 失败: %v", mountURL, err)
	}
}

// DeleteWriteLayer 删除写层目录。
func DeleteWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("删除写层目录 %s 失败: %v", writeURL, err)
	}
}
