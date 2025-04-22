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

