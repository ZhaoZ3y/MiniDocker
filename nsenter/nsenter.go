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
#include <sys/stat.h> // 包含 stat

// 辅助函数：把命令字符串分割成参数数组
char **split_cmd(char *cmd, int *argc) {
    // !! 重要：strtok 会修改字符串，getenv 返回的不能修改，必须复制 !!
    char *cmd_copy = strdup(cmd);
    if (!cmd_copy) {
        fprintf(stderr, "C: split_cmd 无法复制命令字符串: %s\n", strerror(errno));
        *argc = 0;
        return NULL; // 让调用者处理 NULL
    }

    char **argv = NULL;
    char *token = strtok(cmd_copy, " "); // 使用副本
    int count = 0;

    while (token != NULL) {
        argv = realloc(argv, sizeof(char*) * (count + 1));
        if (!argv) {
             fprintf(stderr, "C: split_cmd realloc 失败: %s\n", strerror(errno));
             free(cmd_copy); // 释放副本
             *argc = 0;
             return NULL; // 让调用者处理 NULL
        }
        argv[count] = strdup(token); // !! 重要：复制每个 token !!
        if (!argv[count]) {
            fprintf(stderr, "C: split_cmd 无法复制 token: %s\n", strerror(errno));
            // 清理已分配的 tokens
            for (int j = 0; j < count; j++) free(argv[j]);
            free(argv);
            free(cmd_copy);
            *argc = 0;
            return NULL;
        }
        count++;
        token = strtok(NULL, " ");
    }

    // 最后添加一个 NULL，execvp 需要
    argv = realloc(argv, sizeof(char*) * (count + 1));
    if (!argv) {
        fprintf(stderr, "C: split_cmd realloc (NULL terminator) 失败: %s\n", strerror(errno));
        // 清理已分配的 tokens
        for (int j = 0; j < count; j++) free(argv[j]);
        free(argv);
        free(cmd_copy);
        *argc = 0;
        return NULL;
    }
    argv[count] = NULL;

    *argc = count;
    // free(cmd_copy); // !! 注意：现在不能释放 cmd_copy，因为 argv 中的指针指向它 !!
    // 但因为我们 strdup 了每个 token，现在可以释放 cmd_copy 了
    free(cmd_copy);
    return argv;
}

// 释放 split_cmd 分配的内存
void free_argv(char **argv, int argc) {
    if (!argv) return;
    for (int i = 0; i < argc; i++) {
        free(argv[i]);
    }
    free(argv);
}


// 该函数被标记为 constructor
__attribute__((constructor)) void enter_namespace(void) {
    fprintf(stderr, "C: enter_namespace constructor running in PID: %d\n", getpid()); // 调试

    char *MiniDocker_pid = getenv("MiniDocker_pid");
    // !! 重要：如果 getenv 失败，必须 exit !!
    if (!MiniDocker_pid) {
        //fprintf(stderr, "C: 环境变量 MiniDocker_pid 未设置, 提前返回.\n"); // Debug
        //return; // !! 错误：不能返回，否则 Go 代码会继续执行 !!
        fprintf(stderr, "C: 错误：环境变量 MiniDocker_pid 未设置。\n");
        exit(1); // 必须退出
    }
    fprintf(stderr, "C: 收到 MiniDocker_pid: %s\n", MiniDocker_pid); // 调试

    char *MiniDocker_cmd = getenv("MiniDocker_cmd");
    // !! 重要：如果 getenv 失败，必须 exit !!
    if (!MiniDocker_cmd) {
        //fprintf(stderr, "C: 环境变量 MiniDocker_cmd 未设置, 提前返回.\n"); // Debug
        //return; // !! 错误 !!
        fprintf(stderr, "C: 错误：环境变量 MiniDocker_cmd 未设置。\n");
        exit(1); // 必须退出
    }
     fprintf(stderr, "C: 收到 MiniDocker_cmd: %s\n", MiniDocker_cmd); // 调试

    char *MiniDocker_rootfs = getenv("MiniDocker_rootfs");
    // !! 重要：如果 getenv 失败，必须 exit !!
    if (!MiniDocker_rootfs) {
        fprintf(stderr, "C: 错误: 环境变量 MiniDocker_rootfs 未设置。\n");
        exit(1); // 必须退出
    }
    fprintf(stderr, "C: 收到容器根目录路径: %s\n", MiniDocker_rootfs); // 调试

    int i;
    char nspath[1024];
    char *namespaces[] = { "uts", "ipc", "net", "pid", "mnt" }; // 确保顺序合理

    for (i = 0; i < 5; i++) {
        sprintf(nspath, "/proc/%s/ns/%s", MiniDocker_pid, namespaces[i]);
        int fd = open(nspath, O_RDONLY);
        if (fd < 0) {
            fprintf(stderr, "C: 打开命名空间 %s (%s) 失败: %s\n", namespaces[i], nspath, strerror(errno));
             // 不一定是致命错误，有些容器可能没有全部隔离
             // 但对于 exec，mnt 和 pid 通常是必须的
             if (strcmp(namespaces[i], "mnt") == 0 || strcmp(namespaces[i], "pid") == 0) {
                 exit(1);
             }
             continue; // 尝试继续其他的
        }

        // 使用 setns 进入命名空间
        if (setns(fd, 0) == -1) {
            fprintf(stderr, "C: setns 进入命名空间 %s 失败: %s\n", namespaces[i], strerror(errno));
             close(fd);
             // 同上，mnt 和 pid 失败是致命的
             if (strcmp(namespaces[i], "mnt") == 0 || strcmp(namespaces[i], "pid") == 0) {
                 exit(1);
             }
        }
        close(fd);
        fprintf(stderr, "C: 成功进入命名空间 %s\n", namespaces[i]); // 调试
    }

    // 切换根文件系统
    fprintf(stderr, "C: 尝试 chroot 到 %s\n", MiniDocker_rootfs); // 调试
    if (chroot(MiniDocker_rootfs) != 0) {
        fprintf(stderr, "C: chroot 到 %s 失败: %s\n", MiniDocker_rootfs, strerror(errno));
        exit(1); // chroot 失败是致命的
    }
    fprintf(stderr, "C: chroot 成功\n"); // 调试

    // !! 切换工作目录到新的根目录 "/" !!
    fprintf(stderr, "C: 尝试 chdir 到 \"/\"\n"); // 调试
    if (chdir("/") != 0) {
        fprintf(stderr, "C: chdir 到 \"/\" 失败: %s\n", strerror(errno));
        exit(1); // 切换 CWD 失败是致命的
    }
     // 调试：打印切换后的 CWD
    char cwd_after[1024];
    if (getcwd(cwd_after, sizeof(cwd_after)) != NULL) {
        fprintf(stderr, "C: chdir 后 getcwd() 结果: %s\n", cwd_after);
    } else {
        fprintf(stderr, "C: chdir 后 getcwd() 失败: %s\n", strerror(errno));
    }

    // 挂载 proc (可选但推荐)
    // 检查 /proc 是否存在并且是目录
    struct stat st;
    if (stat("/proc", &st) != 0 || !S_ISDIR(st.st_mode)) {
        fprintf(stderr, "C: /proc 不存在或不是目录, 尝试挂载...\n");
        if (mount("proc", "/proc", "proc", 0, NULL) != 0) {
            // 对于很多命令(如ps)，proc是必须的，挂载失败可能是问题
            fprintf(stderr, "C: 警告: 挂载 /proc 失败: %s\n", strerror(errno));
            // 不退出，但需要注意
        } else {
             fprintf(stderr, "C: 成功挂载 /proc\n");
        }
    } else {
         fprintf(stderr, "C: /proc 已存在\n");
    }


    // 分割命令
    int argc = 0;
    // !! 传递 MiniDocker_cmd 的副本 !!
    char **argv = split_cmd(MiniDocker_cmd, &argc);
    if (argv == NULL || argc == 0) {
        fprintf(stderr, "C: 命令解析失败或命令为空\n");
        exit(1); // 解析失败是致命的
    }

    // 打印将要执行的命令用于调试
    fprintf(stderr, "C: 即将执行命令:");
    for(int k=0; k<argc; k++) {
        fprintf(stderr, " %s", argv[k]);
    }
    fprintf(stderr, "\n");

    // 执行命令
    execvp(argv[0], argv);

    // !! 如果 execvp 返回，说明它失败了 !!
    fprintf(stderr, "C: execvp 执行 %s 失败: %s\n", argv[0], strerror(errno));
    // free_argv(argv, argc); // 在 exit 前理论上可以省略，但好习惯是释放
    exit(1); // execvp 失败必须退出
}
*/
import "C" // 确保 C 代码被编译和链接
