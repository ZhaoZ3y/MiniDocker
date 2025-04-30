package container

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// NewWorkSpace 创建容器的工作空间，包括只读层、写层、挂载点以及用户指定的挂载目录。
// rootURL 是容器工作空间的根目录，mountURL 是容器最终挂载点（容器运行时根目录）。
func NewWorkSpace(volume string, imageName string, containerName string) {
	// 解压 busybox 镜像作为只读层
	CreateReadOnlyLayer(imageName)
	// 创建写层目录，包括 upper 和 work 目录（OverlayFS 结构要求）
	CreateWriteLayer(containerName)
	// 挂载 OverlayFS
	CreateMountPoint(containerName, imageName)
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			MountVolume(volumeURLs, containerName)
			logrus.Infof("挂载宿主机目录 %s", volumeURLs)
		} else {
			logrus.Infof("宿主机目录格式错误，正确格式为 /host/path:/container/path")
		}
	}
}

// volumeUrlExtract 解析用户传入的挂载路径字符串，格式为 /宿主机路径:/容器路径
func volumeUrlExtract(volume string) []string {
	var volumeURLs []string
	volumeURLs = strings.Split(volume, ":")
	return volumeURLs
}

// CreateReadOnlyLayer 创建只读层：解压 busybox.tar 到指定目录。
func CreateReadOnlyLayer(imageName string) error {
	unTarFolderUrl := RootURL + "/" + imageName + "/"
	imageUrl := RootURL + "/" + imageName + ".tar"
	exist, err := PathExists(unTarFolderUrl)
	if err != nil {
		logrus.Errorf("判断目录 %s 是否存在失败: %v", unTarFolderUrl, err)
	}

	// 如果 busybox 目录不存在，创建目录并解压
	if !exist {
		if err := os.MkdirAll(unTarFolderUrl, 0777); err != nil {
			logrus.Errorf("创建目录 %s 失败: %v", unTarFolderUrl, err)
		}
		if _, err := exec.Command("tar", "-xvf", imageUrl, "-C", unTarFolderUrl).CombinedOutput(); err != nil {
			logrus.Errorf("解压目录 %s 失败: %v", unTarFolderUrl, err)
		}
	}
	// 如果 busybox 目录已经存在，直接返回
	// 这里可以添加一些日志或提示信息
	logrus.Infof("只读层 %s 已存在，跳过解压", unTarFolderUrl)
	return nil
}

// CreateWriteLayer 创建写层目录，包括 overlay 所需的 upper 和 work 子目录。
func CreateWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(WriteLayerURL, containerName)
	if err := os.MkdirAll(writeURL, 0777); err != nil && !os.IsExist(err) {
		logrus.Errorf("创建写层目录 %s 失败: %v", writeURL, err)
	}
}

// CreateMountPoint 创建挂载点，并将 OverlayFS 挂载到该目录。
func CreateMountPoint(containerName string, imageName string) bool {
	// 构造挂载相关路径
	lowerDir := filepath.Join(RootURL, imageName)
	upperDir := filepath.Join(RootURL, "writeLayer", containerName, "upper")
	workDir := filepath.Join(RootURL, "writeLayer", containerName, "work")
	mountPoint := filepath.Join(MntURL, containerName)

	// 创建需要的目录
	dirs := []string{lowerDir, upperDir, workDir, mountPoint}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0777); err != nil {
			logrus.Errorf("创建目录 %s 失败: %v", dir, err)
			return false
		}
	}

	// 构造 overlay 挂载参数，并执行 mount 命令
	options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerDir, upperDir, workDir)
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", options, mountPoint)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("挂载 OverlayFS 失败: %v", err)
		return false
	}

	return true
}

// MountVolume 将宿主机目录挂载到容器文件系统的指定路径下
// volumeURLs[0] 为宿主机路径，volumeURLs[1] 为容器内部路径（相对 mountURL）
func MountVolume(volumeURLs []string, containerName string) error {
	parentURL := volumeURLs[0]
	if err := os.MkdirAll(parentURL, 0777); err != nil {
		logrus.Infof("创建宿主机目录 %s 失败: %v", parentURL, err)
	}

	containerURL := volumeURLs[1]
	mntURL := fmt.Sprintf(MntURL, containerName)
	containerVolumeURL := filepath.Join(mntURL, containerURL)

	if err := os.MkdirAll(containerVolumeURL, 0777); err != nil {
		logrus.Infof("创建容器挂载点目录 %s 失败: %v", containerVolumeURL, err)
	}

	// 使用 bind mount 将宿主机目录挂载到容器目录
	cmd := exec.Command("mount", "--bind", parentURL, containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("bind mount 失败: %v", err)
		return err
	}
	return nil
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

// DeleteWorkSpace 删除容器工作空间，包含卸载挂载点和清理写层目录。
// 参数：
// - rootURL：容器文件系统的根路径
// - mountURL：挂载点路径
// - volume：卷配置字符串，如果非空则表示容器挂载了卷
func DeleteWorkSpace(volume string, containerName string) {
	if volume != "" {
		// 如果挂载了卷，解析卷路径
		volumeURLs := volumeUrlExtract(volume)
		// 如果解析结果合法，执行带卷的卸载逻辑
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			DeleteMountPointWithVolume(volumeURLs, containerName)
		} else {
			// 否则仅卸载挂载点
			DeleteMountPoint(containerName)
		}
	} else {
		// 如果没有挂载卷，直接卸载挂载点
		DeleteMountPoint(containerName)
	}
	// 删除写层目录
	DeleteWriteLayer(containerName)
}

// DeleteMountPoint 卸载挂载点并删除挂载点目录。
// 参数：
// - rootURL：容器根路径（本函数中未使用）
// - mountURL：挂载点路径
// DeleteMountPoint 卸载挂载点并删除挂载点目录。
func DeleteMountPoint(containerName string) error {
	mntURL := fmt.Sprintf(MntURL, containerName)
	// 执行 umount 命令卸载挂载点
	_, err := exec.Command("umount", mntURL).CombinedOutput()
	if err != nil {
		logrus.Errorf("卸载挂载点 %s 失败: %v", mntURL, err)
		return err
	}
	// 删除挂载点目录
	if err := os.RemoveAll(mntURL); err != nil {
		logrus.Errorf("删除挂载点目录 %s 失败: %v", mntURL, err)
		return err
	}
	return nil
}

// DeleteMountPointWithVolume 卸载带有数据卷的挂载点，并删除挂载点目录。
// 参数：
// - rootURL：容器根路径（未使用）
// - mountURL：挂载点路径
// - volumeURLs：长度为2的字符串数组，包含宿主机卷路径和容器内挂载路径
func DeleteMountPointWithVolume(volumeURLs []string, containerName string) error {
	// 拼接容器内部卷的完整挂载路径
	mntURL := fmt.Sprintf(MntURL, containerName)
	containerUrl := filepath.Join(mntURL, volumeURLs[1])
	if exist, _ := PathExists(containerUrl); !exist {
		logrus.Warnf("挂载点 %s 不存在，跳过卸载", containerUrl)
		return nil
	}
	// 先卸载容器内部卷的挂载路径
	if _, err := exec.Command("umount", containerUrl).CombinedOutput(); err != nil {
		logrus.Errorf("卸载容器内部卷 %s 失败: %v", containerUrl, err)
		return err
	}
	if _, err := exec.Command("umount", mntURL).CombinedOutput(); err != nil {
		logrus.Errorf("卸载挂载点 %s 失败: %v", mntURL, err)
		return err
	}
	// 删除挂载点目录
	if err := os.RemoveAll(mntURL); err != nil {
		logrus.Infof("删除挂载点目录 %s 失败: %v", mntURL, err)
	}
	return nil
}

// DeleteWriteLayer 删除容器的写层目录。
// 参数：
// - rootURL：容器根路径，写层目录位于 rootURL/writeLayer/
func DeleteWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(WriteLayerURL, containerName)
	if err := os.RemoveAll(writeURL); err != nil {
		logrus.Errorf("删除写层目录 %s 失败: %v", writeURL, err)
	}
}
