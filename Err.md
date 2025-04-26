# 本次写项目时出现的错误
## 1.Windows 11 WSL2 Ubuntu 22.04 挂载Err
我的电脑是windows11，从书上了解到这个项目是要求需要在Linux下运行的，所以我选择了WSL2来运行这个项目。

但是起初觉得VMware的虚拟机太麻烦了，所以我选择了WSL2来运行这个项目。
但是在运行时出现了以下错误：
```shell
<3>WSL (112108) ERROR: No such file or directory @C:/__w/1/s/src/linux/mountutil\mountutilcpp.h:17 (MountEnum)
<3>WSL (112108 - Relay) ERROR: UtilTranslatePathList:2878: Failed to translate C:\Users\Zz3y\AppData\Local\Programs\cursor\resources\app\bin
<3>WSL (112108) ERROR: No such file or directory @C:/__w/1/s/src/linux/mountutil\mountutilcpp.h:17 (MountEnum)
<3>WSL (112108 - Relay) ERROR: UtilTranslatePathList:2878: Failed to translate C:\Program Files\Common Files\Oracle\Java\javapath
<3>WSL (112108) ERROR: No such file or directory @C:/__w/1/s/src/linux/mountutil\mountutilcpp.h:17 (MountEnum)
<3>WSL (112108 - Relay) ERROR: UtilTranslatePathList:2878: Failed to translate C:\Program Files (x86)\Common Files\Oracle\Java\java8path
<3>WSL (112108) ERROR: No such file or directory @C:/__w/1/s/src/linux/mountutil\mountutilcpp.h:17 (MountEnum)
<3>WSL (112108 - Relay) ERROR: UtilTranslatePathList:2878: Failed to translate C:\Program Files (x86)\Common Files\Oracle\Java\javapath
<3>WSL (112108) ERROR: No such file or directory @C:/__w/1/s/src/linux/mountutil\mountutilcpp.h:17 (MountEnum)
<3>WSL (112108 - Relay) ERROR: UtilTranslatePathList:2878: Failed to translate C:\Program Files\SSHFS-Win\bin
<3>WSL (112108) ERROR: No such file or directory @C:/__w/1/s/src/linux/mountutil\mountutilcpp.h:17 (MountEnum)
<3>WSL (112108 - Relay) ERROR: UtilTranslatePathList:2878: Failed to translate C:\Users\Zz3y\.krew\bin
<3>WSL (112108) ERROR: No such file or directory @C:/__w/1/s/src/linux/mountutil\mountutilcpp.h:17 (MountEnum)
<3>WSL (112108 - Relay) ERROR: UtilTranslatePathList:2878: Failed to translate D:\Python\Python312\Scripts\
<3>WSL (112108) ERROR: No such file or directory @C:/__w/1/s/src/linux/mountutil\mountutilcpp.h:17 (MountEnum)
<3>WSL (112108 - Relay) ERROR: UtilTranslatePathList:2878: Failed to translate D:\Python\Python312\
<3>WSL (112108) ERROR: No such file or directory @C:/__w/1/s/src/linux/mountutil\mountutilcpp.h:17 (MountEnum)
<3>WSL (112108 - Relay) ERROR: UtilTranslatePathList:2878: Failed to translate C:\Program Files (x86)\Razer Chroma SDK\bin
````
我发现他会把Windows的配置环境变量也挂载到Linux下，导致了这个错误的出现。然后直接中断了我的项目。

后续我想要尝试解决这个问题，然后上网查找发现其他做这个项目的人都是直接在Linux环境下运行的，所以基本上没有人遇到这个问题，我只能转头去问了ChatGPT，然后他说我的环境变量存在不需要的，这怎么可能存在我不需要的东西

```text
这个是ChatGPT的回答：
✅ 解决方案：
✅ 方法一：清理 Windows 系统环境变量中无效的 PATH 项
打开 Windows 系统设置

搜索并进入「环境变量」

找到你的用户 PATH 和系统 PATH

清理所有不存在的路径，比如：

C:\Users\Zz3y\.krew\bin

D:\Python\Python312\Scripts\

C:\Program Files (x86)\Razer Chroma SDK\bin

关闭再重新打开你的 WSL 终端

✅ 方法二：在 WSL 中避免加载这些路径
你也可以临时绕过它们，不让 WSL 从 Windows 加载环境变量：

编辑你的 .bashrc 或 .zshrc，加上如下内容：

bash
复制
编辑
export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
这会重设 WSL 内的 PATH 变量，不再引入 Windows 的那些错误路径（注意这也会失去使用 Windows 程序的能力）。
```

我windows的环境变量全是重要的怎么可能会有不需要的变量，所以我直接忽略了解决方法1，然后我看到解决方法2，他说如果这样改会导致WSL失去Windows的能力，我感觉有点本末倒置了
我本里就是需要WSL来使用Windows的能力，所以我也直接忽略了这个方法。

最后我只能选择使用VMware的虚拟机来运行这个项目了。麻烦就麻烦吧，起码不会把我的Windows的环境变量挂载到Linux下了，把我的电脑的环境搞成一团遭。

## 2. /proc/self/exe 无法找到
```shell
ERRO[0000] fork/exec /proc/self/exe: no such file or directory
```
在我启动项目的时候出现了这个问题是书本内不存在且没有描述的错误，为此我想了很久没有找到解决方法，最后只能求助于ChatGPT

然后ChatGPT告诉我子进程在尝试通过 /proc/self/exe 重新执行当前程序时失败了。

在`NewParentProcess()`中使用
```go
cmd := exec.Command("/proc/self/exe", "init")
```

在Linux中，`/proc/self/exe`是一个特殊的符号链接，指向当前进程的可执行文件。这个链接在某些情况下可能会失效，尤其是在容器化环境中。
但 在一些特定环境（特别是构建后的二进制文件运行时）会失败，尤其是 go build 输出在临时目录中，或者文件系统有问题。

### 常见原因：可执行文件在不支持的文件系统中
### 解决方法
1. 使用绝对路径替代 /proc/self/exe
```go
selfPath, err := os.Executable()
if err != nil {
logrus.Fatalf("获取自身路径失败: %v", err)
}
cmd := exec.Command(selfPath, "init")
```
已成功解决

## 3.ERRO[0000] 执行 pivot_root 失败: 执行 pivot_root 失败: invalid argument
错误日志
```shell
NFO[0000] 初始化容器                                        
INFO[0000] 用户传入的命令：sh                                   
INFO[0000] 当前工作目录: /home/yzq/Desktop/MiniDocker         
ERRO[0000] 执行 pivot_root 失败: 执行 pivot_root 失败: invalid argument 
INFO[0000] 找到可执行文件路径: /usr/bin/sh   
```

于是我查看书籍发现我的工作目录与书本的不一样，书本是将busybox解压到了busybox目录下并作为工作目录
而我的却是在二进制文件的目录下工作导致失败

后续询问ai给我两个解决方法

### 解决方法1
```go
// 直接使用当前的工作目录进行挂载
func setUpMount() {
	pwd, err := os.Getwd()
	if err != nil {
		logrus.Errorf("获取当前工作目录失败: %v", err)
		return
	}
	logrus.Infof("当前工作目录: %s", pwd)

	// 👇这行代码是关键，强制把当前目录挂载为自己（bind mount）
	if err := syscall.Mount(pwd, pwd, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		logrus.Errorf("绑定当前目录失败: %v", err)
		return
	}

	// 执行 pivot_root
	if err := pivotRoot(pwd); err != nil {
		logrus.Errorf("执行 pivot_root 失败: %v", err)
		return
	}

	// 后面挂载 /proc、/dev 保持不变
	...
}
```

后续尝试之后出现了小问题
```shell
yzq@yzq-virtual-machine:~/Desktop/MiniDocker$ sudo ./MiniDocker run -ti sh
INFO[0000] 用户传入的命令：sh                                   
INFO[0000] 初始化容器                                        
INFO[0000] 当前工作目录: /home/yzq/Desktop/MiniDocker         
ERRO[0000] 挂载 /proc 失败: no such file or directory       
ERRO[0000] 查找路径失败: exec: "sh": executable file not found in $PATH 
2025/04/22 04:34:55 exec: "sh": executable file not found in $PATH
```

接着修复这个小问题
```go
func setUpMount() {
	pwd, err := os.Getwd()
	if err != nil {
		logrus.Errorf("获取当前工作目录失败: %v", err)
		return
	}
	logrus.Infof("当前工作目录: %s", pwd)

	// 👇这行代码是关键，强制把当前目录挂载为自己（bind mount）
	if err := syscall.Mount(pwd, pwd, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		logrus.Errorf("绑定当前目录失败: %v", err)
		return
	}

	// 执行 pivot_root
	if err := pivotRoot(pwd); err != nil {
		logrus.Errorf("执行 pivot_root 失败: %v", err)
		return
	}

	// 后面挂载 /proc、/dev 保持不变
	...
}
```

后续解决失败了
### 解决方法2
书上是在busybox的工作目录下运行

所有我解压了busybox.tar.gz到当前目录下，并且直接将工作目录写死在代码内部

但是也出现了小问题
```shell
yzq@yzq-virtual-machine:~/Desktop/MiniDocker$ sudo ./MiniDocker run -ti sh
INFO[0000] 用户传入的命令：sh                                   
INFO[0000] 初始化容器                                        
INFO[0000] 当前工作目录: /home/yzq/Desktop/MiniDocker/busybox 
ERRO[0000] 挂载 /proc 失败: no such file or directory       
INFO[0000] 找到可执行文件路径: /bin/sh 
```

后续修改在busybox目录下创建了proc和dev目录
```go
// 确保 /proc 目录存在
	procDir := filepath.Join(root, "proc")
	if err := os.MkdirAll(procDir, 0755); err != nil {
		logrus.Errorf("创建 /proc 目录失败: %v", err)
		return
	}
```
后续应该解决了问题，只不过我的好像没有输出全部如同书上的日志，但是ChatGPT说我的是正常的，但愿如此吧，只能继续写了。
### 后日记
后面还是改回去了，我把我的二进制文件放在了root目录下这样就能成功了，主要是后面的内容需要在root环境下才能成功（

## 3. mount: /root/mnt: 未知的文件系统类型“aufs”

错误详情
```shell
mount: /root/mnt: 未知的文件系统类型“aufs”.
ERRO[0000] 挂载失败: exit status 32 
```
后续我进行百度之后发现Ubuntu 22.04（jammy）已经不再默认提供 aufs-tools 包了，因为 AUFS 已经被官方标记为“过时”，推荐使用 overlayfs 替代。

所以我将原来的地方改成了overlayfs而且这个是在Linux内核就支持的无需而外安装

```go
// 使用 overlayfs 替代 aufs
lowerDir := rootURL + "busybox"
upperDir := rootURL + "writeLayer"
workDir := rootURL + "work"
mountPoint := mountURL

_ = os.Mkdir(workDir, 0777) // overlayfs 需要一个专用 work 目录

cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o",
"lowerdir="+lowerDir+",upperdir="+upperDir+",workdir="+workDir,
mountPoint)
```

后续修改挂载的时候也将aufs改成overlay但是ChatGPT推荐另外一个
```go
// MountVolume 挂载宿主机目录到容器挂载点
func MountVolume(rootURL string, mntURL string, volumeURLs []string) {
	// 创建宿主机要挂载的目录
	parentUrl := volumeURLs[0]
	if err := os.MkdirAll(parentUrl, 0777); err != nil {
		log.Infof("创建宿主机目录 %s 失败: %v", parentUrl, err)
	}

	// 在容器挂载点里创建容器内部的挂载目录
	containerUrl := volumeURLs[1]
	containerVolumeURL := mntURL + containerUrl
	if err := os.MkdirAll(containerVolumeURL, 0777); err != nil {
		log.Infof("创建容器内部目录 %s 失败: %v", containerVolumeURL, err)
	}

	// 把宿主机目录挂载到容器内部目录，使用 bind mount
	cmd := exec.Command("mount", "--bind", parentUrl, containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("挂载宿主机目录失败: %v", err)
	}
}

```

## 4.ERRO[0000] 卸载挂载点 /root/mnt 失败: open /dev/null: no such file or directory

很奇怪我的虚拟机内是有这个文件的但是因为某种奇怪的原因没有挂载上似乎是导致了容器在后台运行的时候因为这个错误自动打断了我的容器
后续我直接尝试直接挂载dev目录文件
```go
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

```
```go
// UnmountDev 卸载容器中挂载的 /dev 目录。
func UnmountDev(mountURL string) {
	devPath := filepath.Join(mountURL, "dev")
	if exist, _ := PathExists(devPath); exist {
		cmd := exec.Command("umount", devPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Warnf("卸载容器内 /dev 失败: %v", err)
		} else {
			log.Infof("已卸载容器内 /dev")
		}
	}
}

```
后续还是不行我就在卸载前检查是否存在
```go
func DeleteMountPoint(rootURL string, mountURL string) {
    if exist, _ := PathExists(mountURL); !exist {
        log.Warnf("挂载点 %s 不存在，跳过卸载", mountURL)
        return
    }

    // 卸载 /dev 目录，忽略错误
    UnmountDev(mountURL)

    // 使用不依赖 /dev/null 的方式检查挂载点
    // 使用 cat /proc/mounts 来检查是否挂载，而不是 mountpoint 命令
    cmd := exec.Command("grep", mountURL, "/proc/mounts")
    if output, err := cmd.CombinedOutput(); err != nil || len(output) == 0 {
        log.Infof("挂载点 %s 不是一个有效的挂载点，跳过卸载", mountURL)
        // 直接尝试删除目录
        if err := os.RemoveAll(mountURL); err != nil {
            log.Errorf("删除挂载点目录 %s 失败: %v", mountURL, err)
        }
        return
    }

    // 使用不依赖 /dev/null 的方式检查进程占用
    // 避免使用 lsof 命令，它可能依赖 /dev/null
    // 可以尝试直接强制卸载
    cmd = exec.Command("umount", "-f", mountURL)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        // 如果强制卸载失败，尝试 lazy 卸载
        log.Warnf("强制卸载挂载点 %s 失败，尝试 lazy 卸载", mountURL)
        cmd = exec.Command("umount", "-l", mountURL)
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
            log.Errorf("卸载挂载点 %s 失败: %v", mountURL, err)
            return
        }
    }

    // 删除挂载点目录
    if err := os.RemoveAll(mountURL); err != nil {
        log.Errorf("删除挂载点目录 %s 失败: %v", mountURL, err)
    }
}
```

```go
func DeleteMountPointWithVolume(rootURL string, mountURL string, volumeURLs []string) {
    // 拼接容器内部卷的完整挂载路径
    containerUrl := mountURL + volumeURLs[1]
    if exist, _ := PathExists(containerUrl); !exist {
        log.Warnf("挂载点 %s 不存在，跳过卸载", containerUrl)
        return
    }
    
    // 卸载 /dev 目录
    UnmountDev(mountURL)

    // 先卸载容器内部卷的挂载路径
    cmd := exec.Command("umount", "-f", containerUrl)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        // 尝试 lazy 卸载
        cmd = exec.Command("umount", "-l", containerUrl)
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
            log.Errorf("卸载挂载点 %s 失败: %v", containerUrl, err)
        }
    }
    
    // 再卸载 mountURL 本身
    cmd = exec.Command("umount", "-f", mountURL)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        // 尝试 lazy 卸载
        cmd = exec.Command("umount", "-l", mountURL)
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
            log.Errorf("卸载挂载点 %s 失败: %v", mountURL, err)
        }
    }
    
    // 删除挂载点目录
    if err := os.RemoveAll(mountURL); err != nil {
        log.Infof("删除挂载点目录 %s 失败: %v", mountURL, err)
    }
}
```