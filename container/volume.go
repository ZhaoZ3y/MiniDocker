package container

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// NewWorkSpace 创建容器的工作空间，包括只读层、写层、挂载点以及用户指定的挂载目录。
// rootURL 是容器工作空间的根目录，mountURL 是容器最终挂载点（容器运行时根目录）。
func NewWorkSpace(rootURL string, mountURL string, volume string) {
	// 确保根目录存在
	if err := os.MkdirAll(rootURL, 0777); err != nil && !os.IsExist(err) {
		logrus.Errorf("创建根目录 %s 失败: %v", rootURL, err)
		return
	}
	// 解压 busybox 镜像作为只读层
	CreateReadOnlyLayer(rootURL)
	// 创建写层目录，包括 upper 和 work 目录（OverlayFS 结构要求）
	CreateWriteLayer(rootURL)
	// 创建挂载点目录
	if err := os.MkdirAll(mountURL, 0777); err != nil && !os.IsExist(err) {
		logrus.Errorf("创建挂载点目录 %s 失败: %v", mountURL, err)
		return
	}
	// 如果用户指定了 volume 参数，则解析并挂载宿主机目录
	CreateDevices(mountURL)
	// 使用 OverlayFS 合并只读层和写层，挂载到 mountURL 上
	// 挂载 OverlayFS
	if mountSuccess := CreateMountPoint(rootURL, mountURL); mountSuccess {
		// 挂载成功后再次创建设备文件（确保挂载后设备文件存在）
		CreateDevices(mountURL)
	} else {
		// 处理挂载失败
		logrus.Errorf("OverlayFS挂载失败，无法继续")
		return
	}
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			MountVolume(rootURL, mountURL, volumeURLs)
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

// MountVolume 将宿主机目录挂载到容器文件系统的指定路径下
// volumeURLs[0] 为宿主机路径，volumeURLs[1] 为容器内部路径（相对 mountURL）
func MountVolume(rootURL string, mountURL string, volumeURLs []string) {
	parentURL := volumeURLs[0] // 宿主机目录 /root/volume

	// 确保宿主机目录存在
	if err := os.MkdirAll(parentURL, 0777); err != nil && !os.IsExist(err) {
		logrus.Errorf("创建宿主机目录 %s 失败: %v", parentURL, err)
	}

	containerURL := volumeURLs[1] // 容器内目录路径 /containerVolume

	// 确保容器内目录名是相对路径（不以/开头）
	if strings.HasPrefix(containerURL, "/") {
		containerURL = containerURL[1:] // 去掉开头的斜杠
	}

	containerVolumeURL := filepath.Join(mountURL, containerURL)

	// 创建容器内挂载目录
	if err := os.MkdirAll(containerVolumeURL, 0777); err != nil && !os.IsExist(err) {
		logrus.Errorf("创建容器目录 %s 失败: %v", containerVolumeURL, err)
	}

	// 确保要挂载的宿主机目录是存在且可用的
	fileInfo, err := os.Stat(parentURL)
	if err != nil || !fileInfo.IsDir() {
		logrus.Errorf("宿主机目录 %s 不存在或不是一个目录: %v", parentURL, err)
		return
	}

	// 使用 bind mount 挂载宿主机目录到容器目录
	cmd := exec.Command("mount", "--bind", parentURL, containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("挂载宿主机目录 %s 到容器目录 %s 失败: %v", parentURL, containerVolumeURL, err)
	} else {
		logrus.Infof("成功挂载宿主机目录 %s 到容器目录 %s", parentURL, containerVolumeURL)

		// 验证挂载是否正确
		verifyCmd := exec.Command("findmnt", "--output", "SOURCE,TARGET", containerVolumeURL)
		output, _ := verifyCmd.CombinedOutput()
		logrus.Infof("挂载详情: %s", string(output))
	}
}

// CreateReadOnlyLayer 创建只读层：解压 busybox.tar 到指定目录。
func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := rootURL + "busybox/"       // 解压路径
	busyboxTarURL := rootURL + "busybox.tar" // busybox 压缩包路径

	exist, err := PathExists(busyboxURL)
	if err != nil {
		logrus.Errorf("判断目录 %s 是否存在失败: %v", busyboxURL, err)
	}

	// 如果 busybox 目录不存在，创建目录并解压
	if !exist {
		if err := os.MkdirAll(busyboxURL, 0777); err != nil {
			logrus.Errorf("创建目录 %s 失败: %v", busyboxURL, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			logrus.Errorf("解压目录 %s 失败: %v", busyboxURL, err)
		}
	}
}

// CreateWriteLayer 创建写层目录，包括 overlay 所需的 upper 和 work 子目录。
func CreateWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.MkdirAll(writeURL, 0777); err != nil && !os.IsExist(err) {
		logrus.Errorf("创建写层目录 %s 失败: %v", writeURL, err)
	}
}

func CreateMountPoint(rootURL string, mountURL string) bool {
	if err := os.MkdirAll(mountURL, 0777); err != nil && !os.IsExist(err) {
		logrus.Errorf("创建挂载点目录 %s 失败: %v", mountURL, err)
		return false
	}

	lowerDir := rootURL + "busybox"          // 只读层目录
	upperDir := rootURL + "writeLayer/upper" // 写层 upper 目录
	workDir := rootURL + "writeLayer/work"   // OverlayFS 工作目录

	// 创建 upper 和 work 子目录
	if err := os.MkdirAll(upperDir, 0777); err != nil && !os.IsExist(err) {
		logrus.Errorf("创建 upper 目录失败: %v", err)
		return false
	}
	if err := os.MkdirAll(workDir, 0777); err != nil && !os.IsExist(err) {
		logrus.Errorf("创建 work 目录失败: %v", err)
		return false
	}

	// 构造 overlay 挂载参数，并执行 mount 命令
	options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerDir, upperDir, workDir)
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", options, mountURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("挂载 OverlayFS 失败: %v", err)
		return false
	}

	return true
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

// CreateDevices 创建基本设备文件，包括 /dev/null、/dev/zero、/dev/random 和 /dev/urandom。
func CreateDevices(mountURL string) {
	// 创建 /dev 目录
	devDir := filepath.Join(mountURL, "dev")
	if err := os.MkdirAll(devDir, 0755); err != nil && !os.IsExist(err) {
		logrus.Errorf("创建 /dev 目录失败: %v", err)
		return
	}

	// 创建 /dev/null - 使用 syscall.Mknod 替代 exec.Command
	nullPath := filepath.Join(devDir, "null")
	if err := syscall.Mknod(nullPath, syscall.S_IFCHR|uint32(0666), int(unix.Mkdev(1, 3))); err != nil && !os.IsExist(err) {
		logrus.Errorf("使用 syscall 创建 /dev/null 设备失败: %v", err)
	} else {
		if err := os.Chmod(nullPath, 0666); err != nil {
			logrus.Errorf("修改 /dev/null 权限失败: %v", err)
		}
	}

	// 创建 /dev/zero
	zeroPath := filepath.Join(devDir, "zero")
	if err := syscall.Mknod(zeroPath, syscall.S_IFCHR|uint32(0666), int(unix.Mkdev(1, 5))); err != nil && !os.IsExist(err) {
		logrus.Errorf("使用 syscall 创建 /dev/zero 设备失败: %v", err)
	} else {
		if err := os.Chmod(zeroPath, 0666); err != nil {
			logrus.Errorf("修改 /dev/zero 权限失败: %v", err)
		}
	}

	// 创建 /dev/random
	randomPath := filepath.Join(devDir, "random")
	if err := syscall.Mknod(randomPath, syscall.S_IFCHR|uint32(0666), int(unix.Mkdev(1, 8))); err != nil && !os.IsExist(err) {
		logrus.Errorf("使用 syscall 创建 /dev/random 设备失败: %v", err)
	} else {
		if err := os.Chmod(randomPath, 0666); err != nil {
			logrus.Errorf("修改 /dev/random 权限失败: %v", err)
		}
	}

	// 创建 /dev/urandom
	urandomPath := filepath.Join(devDir, "urandom")
	if err := syscall.Mknod(urandomPath, syscall.S_IFCHR|uint32(0666), int(unix.Mkdev(1, 9))); err != nil && !os.IsExist(err) {
		logrus.Errorf("使用 syscall 创建 /dev/urandom 设备失败: %v", err)
	} else {
		if err := os.Chmod(urandomPath, 0666); err != nil {
			logrus.Errorf("修改 /dev/urandom 权限失败: %v", err)
		}
	}

	logrus.Infof("基本设备文件创建完成")
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
// 参数：
// - rootURL：容器根路径（本函数中未使用）
// - mountURL：挂载点路径
// DeleteMountPoint 卸载挂载点并删除挂载点目录。
func DeleteMountPoint(rootURL string, mountURL string) {
	if exist, _ := PathExists(mountURL); !exist {
		logrus.Warnf("挂载点 %s 不存在，跳过卸载", mountURL)
		return
	}
	// 确保没有进程占用挂载点
	cmd := exec.Command("lsof", "+D", mountURL)
	output, err := cmd.CombinedOutput()
	if err == nil && len(output) > 0 {
		logrus.Warnf("挂载点 %s 被进程占用，无法卸载", mountURL)
		return
	}

	// 执行 umount 命令卸载挂载点
	cmd = exec.Command("umount", mountURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("卸载挂载点 %s 失败: %v", mountURL, err)
		return
	}

	// 删除挂载点目录
	if err := os.RemoveAll(mountURL); err != nil {
		logrus.Errorf("删除挂载点目录 %s 失败: %v", mountURL, err)
	}
}

// DeleteMountPointWithVolume 卸载带有数据卷的挂载点，并删除挂载点目录。
// 参数：
// - rootURL：容器根路径（未使用）
// - mountURL：挂载点路径
// - volumeURLs：长度为2的字符串数组，包含宿主机卷路径和容器内挂载路径
func DeleteMountPointWithVolume(rootURL string, mountURL string, volumeURLs []string) {
	// 拼接容器内部卷的完整挂载路径
	containerUrl := mountURL + volumeURLs[1]
	if exist, _ := PathExists(containerUrl); !exist {
		logrus.Warnf("挂载点 %s 不存在，跳过卸载", containerUrl)
		return
	}
	// 先卸载容器内部卷的挂载路径
	cmd := exec.Command("umount", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("卸载挂载点 %s 失败: %v", containerUrl, err)
	}
	// 再卸载 mountURL 本身
	cmd = exec.Command("umount", mountURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("卸载挂载点 %s 失败: %v", mountURL, err)
	}
	// 删除挂载点目录
	if err := os.RemoveAll(mountURL); err != nil {
		logrus.Infof("删除挂载点目录 %s 失败: %v", mountURL, err)
	}
}

// DeleteWriteLayer 删除容器的写层目录。
// 参数：
// - rootURL：容器根路径，写层目录位于 rootURL/writeLayer/
func DeleteWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.RemoveAll(writeURL); err != nil {
		logrus.Errorf("删除写层目录 %s 失败: %v", writeURL, err)
	}
}
