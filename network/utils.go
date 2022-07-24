/*
@Time : 2022/7/24 18:16
@Author : lianyz
@Description :
*/

package network

import (
	"github.com/sirupsen/logrus"
	"os/exec"
	"strings"
)

func setupIPTables(iptablesCmd string) error {
	logrus.Infof("iptables %s", iptablesCmd)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	output, err := cmd.Output()
	if err != nil {
		logrus.Errorf("iptables output: %v, err: %v", output, err)
		return err
	}
	return nil
}
