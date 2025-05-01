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

    argv = realloc(argv, sizeof(char*) * (count + 1));
    argv[count] = NULL;

    *argc = count;
    return argv;
}

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

    char *MiniDocker_rootfs = getenv("MiniDocker_rootfs");
    if (!MiniDocker_rootfs) {
        fprintf(stderr, "未设置挂载目录路径 MiniDocker_rootfs\n");
        exit(1);
    }

    int i;
    char nspath[1024];
    char *namespaces[] = { "uts", "ipc", "net", "pid", "mnt" };

    // 依次进入各 namespace
    for (i = 0; i < 5; i++) {
        sprintf(nspath, "/proc/%s/ns/%s", MiniDocker_pid, namespaces[i]);
        int fd = open(nspath, O_RDONLY);
        if (fd < 0) {
            fprintf(stderr, "打开命名空间 %s 失败: %s\n", namespaces[i], strerror(errno));
            exit(1);
        }
        if (setns(fd, 0) == -1) {
            fprintf(stderr, "进入命名空间 %s 失败: %s\n", namespaces[i], strerror(errno));
            close(fd);
            exit(1);
        }
        close(fd);
    }

    if (chroot(MiniDocker_rootfs) != 0) {
        fprintf(stderr, "chroot 到容器挂载目录失败: %s\n", strerror(errno));
        exit(1);
    }

    if (chdir("/") != 0) {
        fprintf(stderr, "切换工作目录失败: %s\n", strerror(errno));
        exit(1);
    }

    // 确保 /proc 挂载
    if (access("/proc/self", F_OK) != 0) {
        if (mount("proc", "/proc", "proc", 0, NULL) != 0) {
            fprintf(stderr, "挂载 /proc 失败: %s\n", strerror(errno));
        }
    }

    int argc = 0;
    char **argv = split_cmd(MiniDocker_cmd, &argc);
    if (argv == NULL || argc == 0) {
        fprintf(stderr, "命令解析失败\n");
        exit(1);
    }

    execvp(argv[0], argv);

    fprintf(stderr, "执行命令 %s 失败: %s\n", argv[0], strerror(errno));
    exit(1);
}
*/
import "C"
