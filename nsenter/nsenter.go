package nsenter

/*
#define _GNU_SOURCE
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h>
#include <sys/mount.h>

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
    MiniDocker_pid = getenv("MiniDocker_pid");
    if (!MiniDocker_pid) {
        return;  // 没有设置环境变量，直接返回
    }

    char *MiniDocker_cmd;
    MiniDocker_cmd = getenv("MiniDocker_cmd");
    if (!MiniDocker_cmd) {
        return;  // 没有命令，直接返回
    }

    // 获取容器挂载目录路径
    char *MiniDocker_rootfs = getenv("MiniDocker_rootfs");
    if (!MiniDocker_rootfs) {
        fprintf(stderr, "未设置挂载目录路径 MiniDocker_rootfs\n");
        exit(1);
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

	// 打印切换前的工作目录，用于调试
       char cwd_before[1024];
    if (getcwd(cwd_before, sizeof(cwd_before)) != NULL) {
        fprintf(stderr, "chroot 前工作目录: %s\n", cwd_before);
    }

    // 切换到容器的根文件系统
    if (chroot(MiniDocker_rootfs) != 0) {
        fprintf(stderr, "chroot 到容器挂载目录失败: %s\n", strerror(errno));
        exit(1);
    }

    // 切换工作目录
    if (chdir("/") != 0) {
        fprintf(stderr, "切换工作目录失败: %s\n", strerror(errno));
        exit(1);
    }

    // 打印 chroot 后的工作目录
    char cwd_after[1024];
    if (getcwd(cwd_after, sizeof(cwd_after)) != NULL) {
        fprintf(stderr, "chroot 后工作目录: %s\n", cwd_after);
    }

    // 确保 /proc 在容器内部已挂载
    if (access("/proc/self", F_OK) != 0) {
        if (mount("proc", "/proc", "proc", 0, NULL) != 0) {
            fprintf(stderr, "挂载 /proc 失败: %s\n", strerror(errno));
        }
    }

    // 分割命令字符串为参数数组
    int argc = 0;
    char **argv = split_cmd(MiniDocker_cmd, &argc);
    if (argv == NULL || argc == 0) {
        fprintf(stderr, "命令解析失败\n");
        exit(1);
    }

    // 使用 execvp 执行命令
    // 在执行前再次确认工作目录是根目录
    execvp(argv[0], argv);

    // 如果 execvp 返回，说明执行失败
    fprintf(stderr, "执行命令 %s 失败: %s\n", argv[0], strerror(errno));
    exit(1);
}
*/
import "C"
