/*
@Time : 2022/6/23 23:43
@Author : lianyz
@Description :
*/

package container

import (
	"fmt"
	"github.com/lianyz/mydocker/common"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
)

// LookContainerLog 查看容器内日志信息
func LookContainerLog(containerName string) {
	logFileName := path.Join(common.DefaultContainerInfoPath, containerName, common.ContainerLogFileName)
	file, err := os.Open(logFileName)
	if err != nil {
		logrus.Errorf("open log file, path: %s, err: %v", logFileName, err)
		return
	}

	bs, err := ioutil.ReadAll(file)
	if err != nil {
		logrus.Errorf("read log file, err: %v", err)
		return
	}

	_, _ = fmt.Fprint(os.Stdout, string(bs))

	fmt.Println()
}
