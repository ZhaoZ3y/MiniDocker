package nsenter

/*
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h> // 需要用到 execvp

// 帮助函数：把命令字符串分割成参数数组
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
    // 定义需要进入的命名空间类型
    char *namespaces[] = { "ipc", "uts", "net", "pid", "mnt" };
    // 遍历所有需要进入的 namespace
    for (i = 0; i < 5; i++) {
        // 构造 namespace 文件的路径，例如 /proc/1234/ns/ipc
        sprintf(nspath, "/proc/%s/ns/%s", MiniDocker_pid, namespaces[i]);

        // 打开 namespace 文件，获得文件描述符
        int fd = open(nspath, O_RDONLY);

        // 通过 setns 系统调用进入指定的 namespace
        if (setns(fd, 0) == -1) {
            // 进入失败，可以打印错误信息（这里注释掉了）
        } else {
            // 成功进入对应 namespace
        }

        // 关闭文件描述符，避免资源泄露
        close(fd);
    }

    // 分割命令字符串为参数数组
    int argc = 0;
    char **argv = split_cmd(MiniDocker_cmd, &argc);
    if (argv == NULL || argc == 0) {
        fprintf(stderr, "split_cmd failed\n");
        exit(1);
    }

    // 进入所有 namespace 后，执行传入的命令
    execvp(argv[0], argv);

    // execvp 出错的话，打印错误信息
    perror("execvp");
    // 命令执行失败后直接退出程序
    exit(1);
    return;
}
*/
import "C"
