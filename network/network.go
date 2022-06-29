/*
@Time : 2022/6/28 22:50
@Author : lianyz
@Description :
*/

package network

import (
	"github.com/lianyz/mydocker/container"
	"github.com/vishvananda/netlink"
	"net"
)

var (
	drivers  = map[string]NetworkDriver{}
	networks = map[string]*Network{}
)

// Network 网络
type Network struct {
	Name    string
	IpRange *net.IPNet
	Driver  string
}

// Endpoint 网络端点
type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"dev"`
	IPAddress   net.IP           `json:"ip"`
	MacAddress  net.HardwareAddr `json:"mac"`
	Network     *Network
	PortMapping []string
}

// NetworkDriver 网络驱动接口
type NetworkDriver interface {
	Name() string

	Create(subnet string, name string) (*Network, error)

	Delete(network Network) error

	Connect(network *Network, endpoint *Endpoint) error

	Disconnect(network Network, endpoint *Endpoint) error
}

// Init 初始化网络驱动
func Init() error {

	// todo

	return nil
}

func Connect(net string, info *container.ContainerInfo) error {

	// todo

	return nil
}

func CreateNetwork(driver, subnet, name string) error {

	// todo

	return nil
}
