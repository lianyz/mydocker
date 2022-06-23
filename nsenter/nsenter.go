/*
@Time : 2022/6/21 23:23
@Author : lianyz
@Description :
*/

package nsenter

/*
#define _GNU_SOURCE
#include <unistd.h>
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>

__attribute__((constructor)) void enter_namespace(void) {
    char *docker_pid;
    docker_pid = getenv("docker_pid");
    if (docker_pid) {
        fprintf(stdout, "[cgo]: get docker_pid=%s\n", docker_pid);
    } else {
        fprintf(stdout, "[cgo]: missing docker_pid env skip nsenter\n");
        return;
    }
    char *docker_cmd;
    docker_cmd = getenv("docker_cmd");
    if (docker_cmd) {
        fprintf(stdout, "[cgo]: get docker_cmd=%s\n", docker_cmd);
    } else {
        fprintf(stdout, "[cgo]: mising docker_cmd env skip nsenter\n");
        return;
    }

    int i;
    char nspath[1024];
    char *namespaces[] = { "ipc", "uts", "net", "pid", "mnt" };
    for (i = 0; i < 5; i++) {
        sprintf(nspath, "/proc/%s/ns/%s", docker_pid, namespaces[i]);
        int fd = open(nspath, O_RDONLY);
        // 调用setns系统调用，进入对应的namespace
        if (setns(fd, 0) == -1) {
            fprintf(stdout, "[cgo]: setns on %s namespace failed: %s\n", namespaces[i], strerror(errno));
        } else {
            fprintf(stdout, "[cgo]: setns on %s namespace succeeded\n", namespaces[i]);
        }
        close(fd);
    }
    int res = system(docker_cmd);
    exit(0);
    return;
}
*/
import "C"
