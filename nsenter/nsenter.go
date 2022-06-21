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
        fprintf(stdout, "get docker_pid=%s\n", docker_pid);
    } else {
        fprintf(stdout, "missing docker_pid env skip nsenter");
        return;
    }
}
*/
import "C"
