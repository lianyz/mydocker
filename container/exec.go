/*
@Time : 2022/6/21 23:06
@Author : lianyz
@Description :
*/

package container

// ExecContainer 重新进入容器
// 通过设置环境变量的方式，让C语言写的程序真正运行
// 通过 setns 的系统调用，重新进入到指定的 PID的 namespace 中
func ExecContainer(containerName string, cmdArray []string) {

}
