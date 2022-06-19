/*
@Time : 2022/6/19 10:16
@Author : lianyz
@Description :
*/

package container

import (
	"fmt"
	"github.com/lianyz/mydocker/common"
	"github.com/sirupsen/logrus"
	"os/exec"
	"path"
)

func CommitContainer(imageName, imagePath string) error {
	if imagePath == "" {
		imagePath = common.RootPath
	}

	imageTar := path.Join(imagePath, fmt.Sprintf("%s.tar", imageName))
	_, err := exec.Command("tar", "-czf", imageTar, "-C", common.MntPath, ".").CombinedOutput();
	if err != nil {
		logrus.Errorf("tar container image, file name:%s, err: %v", imageTar, err)
		return err
	}

	return nil
}
