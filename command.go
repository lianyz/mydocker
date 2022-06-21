/*
@Time : 2022/6/16 22:56
@Author : lianyz
@Description :
*/

package main

import (
	"fmt"
	"github.com/lianyz/mydocker/cgroups/subsystem"
	"github.com/lianyz/mydocker/common"
	"github.com/lianyz/mydocker/container"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: "Create a container with namespace and cgroup limit",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		cli.BoolFlag{
			Name:  "ch",
			Usage: "as child process",
		},
		cli.StringFlag{
			Name:  "m",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		cli.StringFlag{
			Name:  "v",
			Usage: "docker volume",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		},
		cli.StringFlag{
			Name:  "name",
			Usage: "docker name",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container args")
		}
		tty := context.Bool("ti")
		detach := context.Bool("d")
		if tty && detach {
			return fmt.Errorf("ti and d parameter can not both provided")
		}

		asChild := context.Bool("ch")
		volume := context.String("v")
		containerName := context.String("name")

		logrus.Infof("args tty:%v aschild:%v", tty, asChild)
		resourceConfig := &subsystem.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuSet:      context.String("cpuset"),
			CpuShare:    context.String("cpushare"),
		}

		// cmdArray 为容器运行后，执行的第一个命令信息
		var cmdArray []string
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}
		Run(cmdArray, tty, asChild, resourceConfig, volume, containerName)
		return nil
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ch",
			Usage: "as child process",
		},
	},
	Action: func(context *cli.Context) error {
		logrus.Infof("begin init come on. args: %v", context.Args())
		asChild := context.Bool("ch")
		logrus.Infof("init come on. args: %v", asChild)
		return container.RunContainerInitProcess(asChild)
	},
}

// 镜像打包
var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "docker commit a container into image",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "c",
			Usage: "export image path",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("missing container name")
		}

		imageName := context.Args().Get(0)
		imagePath := context.String("c")
		return container.CommitContainer(imageName, imagePath)
	},
}

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list all container",
	Action: func(context *cli.Context) error {
		container.ListContainerInfo()
		return nil
	},
}

var execCommand = cli.Command{
	Name:  "exec",
	Usage: "exec a command into container",
	Action: func(context *cli.Context) error {
		// 如果环境变量里面有PID，则什么都不执行
		pid := os.Getenv(common.EnvExecPid)
		if pid != "" {
			logrus.Info("pid callback pid %s, gid: %d", pid, os.Getgid())
			return nil
		}

		if len(context.Args()) < 2 {
			return fmt.Errorf("missing container name or command")
		}

		var cmdArray []string
		for _, arg := range context.Args().Tail() {
			cmdArray = append(cmdArray, arg)
		}

		containerName := context.Args().Get(0)
		container.ExecContainer(containerName, cmdArray)
		return nil
	},
}
