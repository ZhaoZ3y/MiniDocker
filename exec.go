package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// 定义环境变量名称
const ENV_EXEC_PID = "MiniDocker_pid"
const ENV_EXEC_CMD = "MiniDocker_cmd"
const ENV_EXEC_ROOTFS = "MiniDocker_rootfs" // 保持一致

// ExecContainer 函数 (基本不变, 但可以移除 SysProcAttr 中的 Cloneflags)
func ExecContainer(containerName string, comArray []string) {
	pid, err := GetContainerPidByName(containerName) // 确保此函数实现正确
	if err != nil {
		logrus.Errorf("ExecContainer 获取容器 %s 的 PID 失败: %v", containerName, err)
		return
	}

	cmdStr := strings.Join(comArray, " ")
	logrus.Infof("容器的 PID: %s", pid)
	logrus.Infof("要执行的命令: %s", cmdStr)

	// 考虑让根目录路径更灵活，例如从配置或容器信息中读取
	containerRootfs := fmt.Sprintf("/root/mnt/%s", containerName)
	logrus.Infof("容器根文件系统路径: %s", containerRootfs)

	if _, err := os.Stat(containerRootfs); os.IsNotExist(err) {
		logrus.Errorf("容器根文件系统路径 %s 不存在", containerRootfs)
		return
	}

	// 重新执行自身，触发 C constructor
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr // 将标准错误也传递下去，以便看到 C 代码的 fprintf

	// 设置环境变量给 C constructor
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("%s=%s", ENV_EXEC_PID, pid),
		fmt.Sprintf("%s=%s", ENV_EXEC_CMD, cmdStr),
		fmt.Sprintf("%s=%s", ENV_EXEC_ROOTFS, containerRootfs),
	)

	// SysProcAttr: Setctty 和 Setsid 对于交互式 shell 很重要
	// CLONE_NEWNS 可能不再需要，因为 setns 在 C 代码中处理挂载命名空间
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// Cloneflags: syscall.CLONE_NEWNS, // 可以尝试移除
		Setctty: true, // 设置控制终端
		Setsid:  true, // 创建新会话
	}

	// 运行命令。如果 C 代码的 execvp 成功，Run() 不会返回错误。
	// 如果 C 代码 exit(1)，Run() 会返回 ExitError。
	if err := cmd.Run(); err != nil {
		// 这里只应该在 C 代码 exit 非 0 时或启动失败时打印
		if exitErr, ok := err.(*exec.ExitError); ok {
			// C 代码 exit 非 0
			logrus.Debugf("容器命令执行失败，C 代码退出状态: %d", exitErr.ExitCode())
		} else {
			// 其他启动错误
			logrus.Errorf("启动容器命令进程失败: %v", err)
		}
		// 不再打印 "执行容器命令 ... 失败"，因为 C 代码已经打印了具体错误
	}
}
