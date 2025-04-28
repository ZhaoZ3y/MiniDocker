package main

import (
	"MiniDocker/container"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
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

	// 创建一个新的命令：再次执行自己（/proc/self/exe），并传递参数 "exec"
	cmd := exec.Command("/proc/self/exe", "exec")

	// 将当前进程的标准输入输出错误传递给新进程，保持一致
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 设置环境变量，供 nsenter 中的 enter_namespace 使用
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("%s=%s", ENV_EXEC_PID, pid),
		fmt.Sprintf("%s=%s", ENV_EXEC_CMD, cmdStr),
	)

	// 重要: 设置正确的 TTY 参数
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS, // 新的挂载命名空间
		Setctty:    true,                // 设置控制终端
		Setsid:     true,                // 创建新会话
	}

	// 启动新进程，进入容器的 namespace 并执行命令
	if err := cmd.Run(); err != nil {
		logrus.Errorf("执行容器 %s 发生错误 %v", containerName, err)
	}
}

// GetContainerPidByName 根据容器名字获取其 PID
func GetContainerPidByName(containerName string) (string, error) {
	// 拼接出容器对应的配置文件路径
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName

	// 读取容器的配置信息
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}

	var containerInfo container.Info
	// 反序列化 JSON 数据到 containerInfo 结构体
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return "", err
	}

	// 返回容器的 PID
	return containerInfo.Pid, nil
}
