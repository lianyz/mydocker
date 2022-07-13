/*
@Time : 2022/7/3 17:09
@Author : lianyz
@Description :
*/

package network

import (
	"encoding/json"
	"github.com/lianyz/mydocker/common"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"path"
	"strings"
)

// IPAM 网络IP地址的分配与释放
type IPAM struct {
	SubnetAllocatorPath string
	Subnets             *map[string]string
}

var ipAllocator = &IPAM{
	SubnetAllocatorPath: common.DefaultAllocatorPath,
}

// load 从文件里加载对象信息
func (ipam *IPAM) load() error {
	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}

	file, err := os.Open(ipam.SubnetAllocatorPath)
	defer file.Close()
	if err != nil {
		return err
	}
	defer file.Close()

	bs := make([]byte, 2000)
	n, err := file.Read(bs)
	if err != nil {
		return err
	}

	return json.Unmarshal(bs[:n], ipam.Subnets)
}

// dump 将对象信息保存到文件里
func (ipam *IPAM) dump() error {
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	if err := common.Mkdir(ipamConfigFileDir); err != nil {
		return err
	}

	file, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	bs, _ := json.Marshal(ipam.Subnets)

	_, err = file.Write(bs)
	return err
}

// Allocate 从指定的subnet网段中分配IP地址
func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	// 存放网段中地址分配信息的数组
	ipam.Subnets = &map[string]string{}
	// 从文件中加载已经分配的网段信息
	err = ipam.load()
	if err != nil {
		logrus.Errorf("load allocation info, err: %v", err)
	}

	one, size := subnet.Mask.Size()

	logrus.Infof("allocate info: subnet: %v mask: %v one: %v size: %v",
		subnet, subnet.Mask, one, size)

	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(size-one))
	}

	logrus.Infof("allocate ipam subnet: %v", subnet)

	for c := range (*ipam.Subnets)[subnet.String()] {
		if (*ipam.Subnets)[subnet.String()][c] == '0' {
			ipalloc := []byte((*ipam.Subnets)[subnet.String()])
			ipalloc[c] = '1'
			(*ipam.Subnets)[subnet.String()] = string(ipalloc)
			ip = subnet.IP

			logrus.Infof("allocate ip c: %v ip: %v subnet.IP: %v", c, ip, subnet.IP)

			ipBytes := []byte(ip)
			for i := 0; i < len(ipBytes); i++ {
				logrus.Infof("ip[%d]: %d", i, ipBytes[i])
			}

			for t := uint(4); t > 0; t -= 1 {
				before := []byte(ip)[4-t]
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
				after := []byte(ip)[4-t]

				logrus.Infof("allocate ip[%d] %d-%d delta: %d",
					4-t, before, after, uint8(c>>((t-1)*8)))
			}
			ip[3] += 1
			break
		}
	}

	logrus.Infof("allocate ip: %v", ip)

	err = ipam.dump()
	if err != nil {
		logrus.Errorf("allocate ip, dump ipam info, err: %v", err)
		return nil, err
	}

	return
}

// Release 从指定的subnet网段中释放指定的IP地址
func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	ipam.Subnets = &map[string]string{}

	_, subnet, err := net.ParseCIDR(subnet.String())
	if err != nil {
		return err
	}

	err = ipam.load()
	if err != nil {
		logrus.Errorf("dump allocation info, err: %v", err)
		return err
	}

	c := 0
	releaseIP := ipaddr.To4()
	releaseIP[3] -= 1
	for t := uint(4); t > 0; t -= 1 {
		c += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}

	ipalloc := []byte((*ipam.Subnets)[subnet.String()])
	ipalloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipalloc)

	err = ipam.dump()
	if err != nil {
		logrus.Errorf("release ip, dump ipam info, err: %v", err)
		return err
	}

	return nil
}
