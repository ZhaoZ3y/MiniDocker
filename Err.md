# 本次写项目时出现的错误
## 1. /proc/self/exe 无法找到
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

## 常见原因：可执行文件在不支持的文件系统中
## 解决方法
1. 使用绝对路径替代 /proc/self/exe
```go
selfPath, err := os.Executable()
if err != nil {
logrus.Fatalf("获取自身路径失败: %v", err)
}
cmd := exec.Command(selfPath, "init")
```
