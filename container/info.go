/*
@Time : 2022/6/20 22:23
@Author : lianyz
@Description :
*/

package container

import (
	"encoding/json"
	"github.com/lianyz/mydocker/common"
	"github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type ContainerInfo struct {
	Pid         string   `json:"pid"`
	Id          string   `json:"id"`
	Command     string   `json:"command"`
	Name        string   `json:"name"`
	CreateTime  string   `json:"createName"`
	Status      string   `json:"status"`
	Volume      string   `json:"volume"`
	PortMapping []string `json:"portMapping"`
}

func RecordContainerInfo(containerPID int, cmdArray []string, containerName, containerID string) error {
	info := &ContainerInfo{
		Pid:        strconv.Itoa(containerPID),
		Id:         containerID,
		Command:    strings.Join(cmdArray, ""),
		Name:       containerName,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		Status:     common.Running,
	}

	dir := path.Join(common.DefaultContainerInfoPath, containerName)
	if err := common.Mkdir(dir); err != nil {
		logrus.Errorf("mkdir container dir: %s, err: %v", dir, err)
		return err
	}

	fileName := path.Join(dir, common.ContainerInfoFileName)
	file, err := os.Create(fileName)
	if err != nil {
		logrus.Errorf("create config.json, fileName: %s, err: %v", fileName, err)
		return err
	}

	bs, _ := json.Marshal(info)
	_, err = file.WriteString(string(bs))
	if err != nil {
		logrus.Errorf("write config.json, fileName: %s, err: %v", fileName, err)
		return err
	}

	return nil
}

func DeleteContainerInfo(containerName string) {
	dir := path.Join(common.DefaultContainerInfoPath, containerName)
	err := os.RemoveAll(dir)
	if err != nil {
		logrus.Errorf("remove container info, err: %v", err)
	}
}

func GenContainerID(n int) string {
	letterBytes := "0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
