# MiniDocker
一个简单的MiniDocker，参考[《自己动手写Docker》](https://github.com/ZhaoZ3y/MiniDocker/blob/main/docs/%E8%87%AA%E5%B7%B1%E5%8A%A8%E6%89%8B%E5%86%99Docker.pdf)进行书写，收获良多。
## 环境
我是Windows用户在这个上面吃了很多亏，推荐使用Linux进行书写
WSL2.0、Ubuntu22.04虚拟机

## 功能
1. 实现了容器的创建、删除、前后台运行、列出所有容器、进入处于后台运行的容器等功能
2. 镜像的创建、启动
3. 实现了容器的网络连接

## 如何实现

在终端内执行
```shell
git clone https://github.com/ZhaoZ3y/MiniDocker.git
```
下载项目源码后，通过如下命令拉取依赖源

```shell
go mod tidy
go build .
```
创建好二进制文件后，在Linux环境下运行

## 项目出现的偶现问题与解决方案

本次项目实践出现的问题全部在[错误文档](https://github.com/ZhaoZ3y/MiniDocker/blob/main/docs/Err.md)内已经书写出来了
