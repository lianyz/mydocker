/*
@Time : 2022/6/21 23:13
@Author : lianyz
@Description :
*/

package main

/*
#define _GNU_SOURCE
#include <unistd.h>
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>

*/
import "C"
import "fmt"

func main() {
	fmt.Printf("Hello, cgo, random: %v\n", Random())
}

func Random() int {
	return int(C.random())
}

func Seed(i int) {
	C.srandom(C.uint(i))
}
