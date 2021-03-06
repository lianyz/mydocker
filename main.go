/*
@Time : 2022/6/13 18:49
@Author : lianyz
@Description :
*/

package main

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"

	_ "github.com/lianyz/mydocker/nsenter"
)

const usage = `mydocker`

func main() {
	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = usage

	app.Commands = []cli.Command{
		initCommand,
		runCommand,
		commitCommand,
		listCommand,
		execCommand,
		logCommand,
		stopCommand,
		removeCommand,
		networkCommand,
	}

	app.Before = func(context *cli.Context) error {
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
