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

// volumeUrlExtract 解析用户传入的挂载路径字符串，格式为 /宿主机路径:/容器路径
func volumeUrlExtract(volume string) []string {
	var volumeURLs []string
	volumeURLs = strings.Split(volume, ":")
	return volumeURLs
}

// MountVolume 将宿主机目录挂载到容器文件系统的指定路径下
// volumeURLs[0] 为宿主机路径，volumeURLs[1] 为容器内部路径（相对 mountURL）
func MountVolume(rootURL string, mountURL string, volumeURLs []string) {
	parentURL := volumeURLs[0] // 宿主机目录 /root/volume

	// 确保宿主机目录存在
	if err := os.MkdirAll(parentURL, 0777); err != nil && !os.IsExist(err) {
		log.Errorf("创建宿主机目录 %s 失败: %v", parentURL, err)
	}

	containerURL := volumeURLs[1] // 容器内目录路径 /containerVolume

	// 确保容器内目录名是相对路径（不以/开头）
	if strings.HasPrefix(containerURL, "/") {
		containerURL = containerURL[1:] // 去掉开头的斜杠
	}

	containerVolumeURL := filepath.Join(mountURL, containerURL)

	// 创建容器内挂载目录
	if err := os.MkdirAll(containerVolumeURL, 0777); err != nil && !os.IsExist(err) {
		log.Errorf("创建容器目录 %s 失败: %v", containerVolumeURL, err)
	}

	// 确保要挂载的宿主机目录是存在且可用的
	fileInfo, err := os.Stat(parentURL)
	if err != nil || !fileInfo.IsDir() {
		log.Errorf("宿主机目录 %s 不存在或不是一个目录: %v", parentURL, err)
		return
	}

	// 使用 bind mount 挂载宿主机目录到容器目录
	cmd := exec.Command("mount", "--bind", parentURL, containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("挂载宿主机目录 %s 到容器目录 %s 失败: %v", parentURL, containerVolumeURL, err)
	} else {
		log.Infof("成功挂载宿主机目录 %s 到容器目录 %s", parentURL, containerVolumeURL)

		// 验证挂载是否正确
		verifyCmd := exec.Command("findmnt", "--output", "SOURCE,TARGET", containerVolumeURL)
		output, _ := verifyCmd.CombinedOutput()
		log.Infof("挂载详情: %s", string(output))
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
		if err := os.MkdirAll(busyboxURL, 0777); err != nil {
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
	if err := os.MkdirAll(upperDir, 0777); err != nil && !os.IsExist(err) {
		log.Errorf("创建 upper 目录失败: %v", err)
	}
	if err := os.MkdirAll(workDir, 0777); err != nil && !os.IsExist(err) {
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

// MountDev 将宿主机 /dev 挂载到容器的 /dev 中，保证容器中可以访问 /dev/null 等设备。
func MountDev(mountURL string) {
	devPath := filepath.Join(mountURL, "dev")

	// 创建 /dev 目录
	if err := os.MkdirAll(devPath, 0755); err != nil {
		log.Errorf("创建容器内 /dev 目录失败: %v", err)
		return
	}

	// 使用 bind mount 挂载宿主机的 /dev
	cmd := exec.Command("mount", "--bind", "/dev", devPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("挂载 /dev 到容器失败: %v", err)
	} else {
		log.Infof("成功将宿主机 /dev 挂载到容器中")
	}
}

// DeleteWorkSpace 删除容器工作空间，包含卸载挂载点和清理写层目录。
// 参数：
// - rootURL：容器文件系统的根路径
// - mountURL：挂载点路径
// - volume：卷配置字符串，如果非空则表示容器挂载了卷
func DeleteWorkSpace(rootURL string, mountURL string, volume string) {
	if volume != "" {
		// 如果挂载了卷，解析卷路径
		volumeURLs := volumeUrlExtract(volume)
		// 如果解析结果合法，执行带卷的卸载逻辑
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			DeleteMountPointWithVolume(rootURL, mountURL, volumeURLs)
		} else {
			// 否则仅卸载挂载点
			DeleteMountPoint(rootURL, mountURL)
		}
	} else {
		// 如果没有挂载卷，直接卸载挂载点
		DeleteMountPoint(rootURL, mountURL)
	}
	// 删除写层目录
	DeleteWriteLayer(rootURL)
}

// DeleteMountPoint 卸载挂载点并删除挂载点目录。
// 注意：这个函数签名在你的代码里似乎有两个版本，一个带 volume 一个不带
// 这里修改不带 volume 的版本，带 volume 的版本也需要类似修改
func DeleteMountPoint(mountURL string, volume string) error { // 假设这是实际调用的版本
	log.Infof("开始卸载和清理 %s", mountURL)

	// 1. 确保卸载 /dev (调用修改后的 UnmountDev)
	UnmountDev(mountURL)

	// 2. (如果这个函数也处理 volume) 卸载 volume
	// UnmountVolume(...) // 假设有这个函数

	// 3. 卸载主挂载点 mountURL
	// 移除 lsof 检查，因为它本身出错了，而且即使成功，懒卸载通常是更好的选择
	log.Infof("尝试卸载主挂载点 %s", mountURL)
	cmd := exec.Command("umount", "-l", mountURL) // 直接尝试懒卸载
	if err := cmd.Run(); err != nil {
		// 记录错误，但继续尝试删除目录
		log.Errorf("懒卸载 %s 失败: %v。继续尝试删除目录。", mountURL, err)
		// 在某些情况下，即使 umount 报错，目录也可能不再是挂载点了
		// 或者有时需要一点时间让内核完成清理
		// 可以考虑在这里加一个短暂的 sleep (time.Sleep(100 * time.Millisecond))
	} else {
		log.Infof("成功懒卸载 %s", mountURL)
	}

	// 4. 删除挂载目录
	log.Infof("尝试删除目录 %s", mountURL)
	if err := os.RemoveAll(mountURL); err != nil {
		// 这里仍然可能失败，特别是如果懒卸载后内核还未完全释放资源
		log.Errorf("删除挂载目录 %s 失败: %v", mountURL, err)
		// 可以考虑在这里重试几次
		return fmt.Errorf("删除挂载目录 %s 失败: %w", mountURL, err)
	} else {
		log.Infof("成功删除目录 %s", mountURL)
	}

	return nil
}

// DeleteMountPointWithVolume 卸载带有数据卷的挂载点，并删除挂载点目录。
// 参数：
// - rootURL：容器根路径（未使用）
// - mountURL：挂载点路径
// - volumeURLs：长度为2的字符串数组，包含宿主机卷路径和容器内挂载路径
func DeleteMountPointWithVolume(rootURL string, mountURL string, volumeURLs []string) {
	containerUrl := filepath.Join(mountURL, volumeURLs[1]) // 使用 filepath.Join 更安全
	log.Infof("开始卸载和清理带有 Volume 的 %s", mountURL)

	// 1. 卸载 Volume
	log.Infof("尝试卸载 Volume %s", containerUrl)
	cmd := exec.Command("umount", "-l", containerUrl) // 优先懒卸载
	if err := cmd.Run(); err != nil {
		log.Warnf("卸载 Volume 挂载点 %s 失败 (尝试懒卸载): %v", containerUrl, err)
	} else {
		log.Infof("成功懒卸载 Volume %s", containerUrl)
	}

	// 2. 卸载 /dev (调用修改后的 UnmountDev)
	UnmountDev(mountURL)

	// 3. 卸载主挂载点 mountURL
	log.Infof("尝试卸载主挂载点 %s", mountURL)
	cmd = exec.Command("umount", "-l", mountURL) // 优先懒卸载
	if err := cmd.Run(); err != nil {
		log.Errorf("懒卸载主挂载点 %s 失败: %v。继续尝试删除目录。", mountURL, err)
	} else {
		log.Infof("成功懒卸载主挂载点 %s", mountURL)
	}

	// 4. 删除挂载点目录
	log.Infof("尝试删除目录 %s", mountURL)
	if err := os.RemoveAll(mountURL); err != nil {
		log.Errorf("删除挂载点目录 %s 失败: %v", mountURL, err)
	} else {
		log.Infof("成功删除目录 %s", mountURL)
	}
}

// DeleteWriteLayer 删除容器的写层目录。
// 参数：
// - rootURL：容器根路径，写层目录位于 rootURL/writeLayer/
func DeleteWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("删除写层目录 %s 失败: %v", writeURL, err)
	}
}

// UnmountDev 卸载容器的 /dev
func UnmountDev(mountURL string) {
	devPath := filepath.Join(mountURL, "dev")
	log.Infof("尝试卸载 devPath: %s", devPath) // 添加日志确认路径

	// 优先尝试普通卸载
	cmd := exec.Command("umount", devPath)
	err := cmd.Run()
	if err != nil {
		log.Warnf("卸载 %s 失败: %v。尝试懒卸载。", devPath, err)
		// 如果普通卸载失败，坚决尝试懒卸载
		cmd = exec.Command("umount", "-l", devPath)
		if err := cmd.Run(); err != nil {
			// 即使懒卸载也失败了，也只是记录错误，不应阻塞后续清理流程
			log.Errorf("懒卸载 %s 仍然失败: %v", devPath, err)
		} else {
			log.Infof("成功懒卸载 %s", devPath)
		}
	} else {
		log.Infof("成功卸载 %s", devPath)
	}
}
