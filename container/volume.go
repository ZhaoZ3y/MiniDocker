package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// NewWorkSpace 创建容器的工作空间，包括只读层、写层、挂载点以及用户指定的挂载目录。
// rootURL 是容器工作空间的根目录，mountURL 是容器最终挂载点（容器运行时根目录）。
func NewWorkSpace(rootURL string, mountURL string, volume string) {
	// 解压 busybox 镜像作为只读层
	CreateReadOnlyLayer(rootURL)
	// 创建写层目录，包括 upper 和 work 目录（OverlayFS 结构要求）
	CreateWriteLayer(rootURL)
	// 使用 OverlayFS 合并只读层和写层，挂载到 mountURL 上
	CreateMountPoint(rootURL, mountURL)
	// 挂载宿主机的 /dev 目录到容器的 /dev 目录
	MountDev(mountURL)
	// 如果用户指定了 volume 参数，则解析并挂载宿主机目录
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			MountVolume(rootURL, mountURL, volumeURLs)
			log.Infof("挂载宿主机目录 %s", volumeURLs)
		} else {
			log.Infof("宿主机目录格式错误，正确格式为 /host/path:/container/path")
		}
	}
}

// DeleteWorkSpace 删除容器工作空间
func DeleteWorkSpace(rootURL, mountURL, volume string) {
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			DeleteMountPointWithVolume(rootURL, mountURL, volumeURLs)
		} else {
			DeleteMountPoint(rootURL, mountURL)
		}
	} else {
		DeleteMountPoint(rootURL, mountURL)
	}
	DeleteWriteLayer(rootURL)
}

// CreateReadOnlyLayer 创建只读层：解压 busybox.tar 到指定目录。
func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := filepath.Join(rootURL, "busybox")
	busyboxTarURL := filepath.Join(rootURL, "busybox.tar")

	exist, err := PathExists(busyboxURL)
	if err != nil {
		log.Errorf("检查 busybox 目录失败: %v", err)
	}

	// 如果 busybox 目录不存在，创建目录并解压
	if !exist {
		if err := os.MkdirAll(busyboxURL, 0777); err != nil {
			log.Errorf("创建 busybox 目录失败: %v", err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			log.Errorf("解压 busybox 失败: %v", err)
		}
	}
}

// CreateWriteLayer 创建写层目录，包括 overlay 所需的 upper 和 work 子目录。
func CreateWriteLayer(rootURL string) {
	writeURL := filepath.Join(rootURL, "writeLayer")
	if err := os.MkdirAll(writeURL, 0777); err != nil && !os.IsExist(err) {
		log.Errorf("创建写层目录 %s 失败: %v", writeURL, err)
	}
}

// CreateMountPoint 创建挂载点并挂载 overlay
func CreateMountPoint(rootURL, mountURL string) {
	if err := os.MkdirAll(mountURL, 0777); err != nil && !os.IsExist(err) {
		log.Errorf("创建挂载点目录 %s 失败: %v", mountURL, err)
	}

	lowerDir := filepath.Join(rootURL, "busybox")
	upperDir := filepath.Join(rootURL, "writeLayer", "upper")
	workDir := filepath.Join(rootURL, "writeLayer", "work")

	for _, dir := range []string{upperDir, workDir} {
		if err := os.MkdirAll(dir, 0777); err != nil && !os.IsExist(err) {
			log.Errorf("创建目录 %s 失败: %v", dir, err)
		}
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

// MountVolume 宿主机目录挂载到容器
func MountVolume(rootURL, mountURL string, volumeURLs []string) {
	parentURL := volumeURLs[0]

	if err := os.MkdirAll(parentURL, 0777); err != nil {
		log.Errorf("创建宿主机目录失败: %v", err)
	}

	containerURL := strings.TrimPrefix(volumeURLs[1], "/")
	containerVolumeURL := filepath.Join(mountURL, containerURL)

	if err := os.MkdirAll(containerVolumeURL, 0777); err != nil {
		log.Errorf("创建容器目录失败: %v", err)
	}

	cmd := exec.Command("mount", "--bind", parentURL, containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("挂载宿主机目录失败: %v", err)
	} else {
		log.Infof("成功挂载宿主机目录 %s 到容器目录 %s", parentURL, containerVolumeURL)
	}
}

// DeleteMountPoint 删除挂载点
func DeleteMountPoint(rootURL, mountURL string) {
	UnmountDev(mountURL)

	if IsMounted(mountURL) {
		cmd := exec.Command("umount", mountURL)
		if err := cmd.Run(); err != nil {
			log.Warnf("普通卸载失败，尝试懒卸载: %v", err)
			cmd = exec.Command("umount", "-l", mountURL)
			if err := cmd.Run(); err != nil {
				log.Errorf("懒卸载失败: %v", err)
			}
		}
	}

	if err := os.RemoveAll(mountURL); err != nil {
		log.Errorf("删除挂载点目录失败: %v", err)
	}
}

// DeleteMountPointWithVolume 删除带数据卷的挂载点
func DeleteMountPointWithVolume(rootURL, mountURL string, volumeURLs []string) {
	containerUrl := filepath.Join(mountURL, strings.TrimPrefix(volumeURLs[1], "/"))

	if exist, _ := PathExists(containerUrl); exist && IsMounted(containerUrl) {
		cmd := exec.Command("umount", containerUrl)
		if err := cmd.Run(); err != nil {
			log.Warnf("卸载数据卷失败: %v", err)
			cmd = exec.Command("umount", "-l", containerUrl)
			_ = cmd.Run()
		}
	}

	DeleteMountPoint(rootURL, mountURL)
}

// DeleteWriteLayer 删除写层目录
func DeleteWriteLayer(rootURL string) {
	writeURL := filepath.Join(rootURL, "writeLayer")
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("删除写层目录失败: %v", err)
	}
}

// MountDev 将宿主机 /dev 绑定到容器
func MountDev(mountURL string) {
	devPath := filepath.Join(mountURL, "dev")
	if err := os.MkdirAll(devPath, 0755); err != nil {
		log.Errorf("创建容器内 /dev 目录失败: %v", err)
		return
	}

	cmd := exec.Command("mount", "--bind", "/dev", devPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("挂载 /dev 失败: %v", err)
	} else {
		log.Infof("成功挂载宿主机 /dev 到容器")
	}
}

// UnmountDev 卸载容器的 /dev
func UnmountDev(mountURL string) {
	devPath := filepath.Join(mountURL, "dev")

	if !IsMounted(devPath) {
		log.Infof("%s 未挂载，跳过卸载", devPath)
		return
	}

	cmd := exec.Command("umount", devPath)
	if err := cmd.Run(); err != nil {
		log.Warnf("卸载 %s 失败，尝试懒卸载: %v", devPath, err)
		cmd = exec.Command("umount", "-l", devPath)
		if err := cmd.Run(); err != nil {
			log.Errorf("懒卸载 %s 仍失败: %v", devPath, err)
		}
	}
}

// PathExists 判断路径是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// volumeUrlExtract 解析 volume 参数
func volumeUrlExtract(volume string) []string {
	return strings.Split(volume, ":")
}

// IsMounted 检查路径是否已挂载
func IsMounted(path string) bool {
	cmd := exec.Command("mountpoint", "-q", path)
	return cmd.Run() == nil
}
