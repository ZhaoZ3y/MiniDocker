package nsenter

/*
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h> // 需要用到 execvp

// 辅助函数：把命令字符串分割成参数数组
char **split_cmd(char *cmd, int *argc) {
    char **argv = NULL;
    char *token = strtok(cmd, " ");
    int count = 0;

    while (token != NULL) {
        argv = realloc(argv, sizeof(char*) * (count + 1));
        argv[count] = token;
        count++;
        token = strtok(NULL, " ");
    }

    // 最后添加一个 NULL，execvp 需要
    argv = realloc(argv, sizeof(char*) * (count + 1));
    argv[count] = NULL;

    *argc = count;
    return argv;
}

// 该函数被标记为 constructor，意思是：在 Go 程序加载这个包时，
// 这段 C 代码会自动执行，不需要手动调用。
__attribute__((constructor)) void enter_namespace(void) {
    char *MiniDocker_pid;
    // 获取环境变量 MiniDocker_pid，该环境变量保存的是目标进程的 PID
    MiniDocker_pid = getenv("MiniDocker_pid");
    if (MiniDocker_pid) {
        // 成功获取到 MiniDocker_pid，准备进入对应进程的 namespace
    } else {
        // 如果没有 MiniDocker_pid 环境变量，说明不需要 nsenter，直接返回
        return;
    }

    char *MiniDocker_cmd;
    // 获取环境变量 MiniDocker_cmd，该环境变量保存的是需要执行的命令
    MiniDocker_cmd = getenv("MiniDocker_cmd");
    if (MiniDocker_cmd) {
        // 成功获取到命令，稍后执行
    } else {
        // 如果没有 MiniDocker_cmd，说明没有需要执行的命令，直接返回
        return;
    }

    int i;
    char nspath[1024];
    // 顺序很重要：先进入 uts, ipc, net，再进入 pid，最后进入 mnt
    char *namespaces[] = { "uts", "ipc", "net", "pid", "mnt" };

    // 遍历所有需要进入的 namespace
    for (i = 0; i < 5; i++) {
        // 构造 namespace 文件的路径，例如 /proc/1234/ns/ipc
        sprintf(nspath, "/proc/%s/ns/%s", MiniDocker_pid, namespaces[i]);

        // 打开 namespace 文件，获得文件描述符
        int fd = open(nspath, O_RDONLY);
        if (fd < 0) {
            fprintf(stderr, "打开命名空间 %s 失败: %s\n", namespaces[i], strerror(errno));
            exit(1);
        }

        // 通过 setns 系统调用进入指定的 namespace
        if (setns(fd, 0) == -1) {
            fprintf(stderr, "进入命名空间 %s 失败: %s\n", namespaces[i], strerror(errno));
            close(fd);
            exit(1);
        }
        close(fd);
    }

    // 获取容器根目录路径并切换到容器的文件系统
    char rootfs_path[1024];
    sprintf(rootfs_path, "/proc/%s/root", MiniDocker_pid);

    // 切换到容器的根文件系统
    if (chroot(rootfs_path) != 0) {
        fprintf(stderr, "chroot 到容器根目录失败: %s\n", strerror(errno));
        exit(1);
    }

    // 切换工作目录
    if (chdir("/") != 0) {
        fprintf(stderr, "切换工作目录失败: %s\n", strerror(errno));
        exit(1);
    }

    // 确保 /proc 在容器内部已挂载
    if (access("/proc/self", F_OK) != 0) {
        // 如果 /proc 不存在或无法访问，尝试挂载
        if (mount("proc", "/proc", "proc", 0, NULL) != 0) {
            fprintf(stderr, "挂载 /proc 失败: %s\n", strerror(errno));
            // 这里不退出，因为某些容器可能有特殊配置
        }
    }

    // 分割命令字符串为参数数组
    int argc = 0;
    char **argv = split_cmd(MiniDocker_cmd, &argc);
    if (argv == NULL || argc == 0) {
        fprintf(stderr, "split_cmd failed\n");
        exit(1);
    }

    // 使用 execvp 执行命令
    execvp(argv[0], argv);

    // 如果 execvp 返回，说明执行失败
    fprintf(stderr, "执行命令 %s 失败: %s\n", argv[0], strerror(errno));
    exit(1);
}
*/
import "C"
