package main

import (
	_ "MiniDocker/nsenter" // 引入 nsenter 包，自动执行其中的 C 代码
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// 定义环境变量名称
const ENV_EXEC_PID = "MiniDocker_pid" // 要进入的目标容器进程 PID
const ENV_EXEC_CMD = "MiniDocker_cmd" // 要在容器中执行的命令

// ExecContainer 用于在指定容器内执行命令
func ExecContainer(containerName string, comArray []string) {
	// 通过容器名查找对应的 PID
	pid, err := GetContainerPidByName(containerName)
	if err != nil {
		logrus.Errorf("ExecContainer getContainerPidByName %s 发生错误 %v", containerName, err)
		return
	}

	// 将用户输入的命令数组转成空格分隔的字符串，比如 ["ls", "-l"] -> "ls -l"
	cmdStr := strings.Join(comArray, " ")
	logrus.Infof("容器的 PID: %s", pid)
	logrus.Infof("要执行的命令: %s", cmdStr)

	// 构建容器根文件系统的完整路径
	containerRootfs := fmt.Sprintf("/root/mnt/%s", containerName)
	logrus.Infof("容器根文件系统路径: %s", containerRootfs)

	// 验证容器根文件系统路径是否存在
	if _, err := os.Stat(containerRootfs); os.IsNotExist(err) {
		logrus.Errorf("容器根文件系统路径不存在: %s", containerRootfs)
		return
	}

	// 创建一个新的命令：再次执行自己（/proc/self/exe），并传递参数 "exec"
	cmd := exec.Command("/proc/self/exe", "exec")
	// 将当前进程的标准输入输出错误传递给新进程，保持一致
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	os.Setenv(ENV_EXEC_PID, pid)    // 设置环境变量，供 nsenter 中的 enter_namespace 使用
	os.Setenv(ENV_EXEC_CMD, cmdStr) // 设置环境变量，供 nsenter 中的 enter_namespace 使用

	containerEnv, err := getEnvsByPid(pid)
	if err != nil {
		logrus.Errorf("获取容器环境变量失败: %v", err)
		return
	}

	// 设置环境变量，供 nsenter 中的 enter_namespace 使用
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("%s=%s", ENV_EXEC_PID, pid),
		fmt.Sprintf("%s=%s", ENV_EXEC_CMD, cmdStr),
		fmt.Sprintf("MiniDocker_env=%s", containerEnv),
		fmt.Sprintf("MiniDocker_rootfs=%s", containerRootfs),
	)

	// 重要: 设置正确的 TTY 参数
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS, // 新的挂载命名空间
		Setsid:     true,                // 创建新会话
	}

	// 启动新进程，进入容器的 namespace 并执行命令
	if err := cmd.Run(); err != nil {
		logrus.Errorf("执行容器 %s 发生错误 %v", containerName, err)
	}
}
