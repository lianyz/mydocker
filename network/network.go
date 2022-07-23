/*
@Time : 2022/6/28 22:50
@Author : lianyz
@Description :
*/

package network

import (
	"encoding/json"
	"fmt"
	"github.com/lianyz/mydocker/common"
	"github.com/lianyz/mydocker/container"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"
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
	// Name 驱动名
	Name() string

	// Create 创建网络
	Create(subnet string, name string) (*Network, error)

	// Delete 删除网络
	Delete(network Network) error

	// Connect 连接容器网络端点到网络
	Connect(network *Network, endpoint *Endpoint) error

	// Disconnect 从网络上移除容器网络端点
	Disconnect(network Network, endpoint *Endpoint) error
}

// Init 初始化网络驱动
func Init() error {
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver

	if err := common.Mkdir(common.DefaultNetworkPath); err != nil {
		return err
	}

	// 递归遍历目录
	err := filepath.Walk(common.DefaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		if strings.HasSuffix(nwPath, "/") {
			return nil
		}
		_, nwName := path.Split(nwPath)
		nw := &Network{
			Name: nwName,
		}

		if err := nw.load(nwPath); err != nil {
			logrus.Errorf("error load network: %v", err)
		}

		networks[nwName] = nw

		logrus.Infof("list network, name: %s, ipRange: %v, gatewayIP:%s, driver:%s",
			nwName, nw.IpRange, nw.IpRange.IP, nw.Driver)

		return nil
	})

	if err != nil {
		logrus.Errorf("file path walk, err: %v", err)
		return err
	}

	return nil
}

func CreateNetwork(driverName, subnet, name string) error {

	logrus.Infof("create network %s dirver: %s subnet:%s", name, driverName, subnet)

	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		logrus.Errorf("parse cidr, err: %v", err)
		return err
	}

	logrus.Infof("create network ipNet: %v subnet: %s", ipNet, subnet)

	// 通过IPAM分配网关IP，获取到网段中的第一个IP作为网关的IP
	gatewayIp, err := ipAllocator.Allocate(ipNet)
	if err != nil {
		logrus.Errorf("allocate ip, err: %v", err)
	}

	ipNet.IP = gatewayIp

	logrus.Infof("create network, ipNet: %v getwayIp: %v", ipNet, gatewayIp)

	// 创建网络
	logrus.Infof("dirvers: %v ipNet: %v driver: %v", drivers, ipNet, driverName)
	driver := drivers[driverName]
	if driver == nil {
		err := fmt.Errorf("can not find driver %s", driverName)
		return err
	}
	nw, err := drivers[driverName].Create(ipNet.String(), name)
	if err != nil {
		return err
	}

	// 将对象保存到文件中
	err = nw.dump(common.DefaultNetworkPath)
	if err != nil {
		logrus.Errorf("dump network, err: %v", err)
		return err
	}

	return nil
}

func AllocateIP(networkName string) error {
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no such network: %s", networkName)
	}

	logrus.Infof("network: name: %s ipRange: %v driver: %v",
		network.Name, network.IpRange, network.Driver)

	// 分配容器IP地址,输入参数不能直接用network.IpRange，否则输出的ip地址为v6格式
	ip, err := ipAllocator.Allocate(network.IpRange)
	if err != nil {
		return err
	}

	logrus.Infof("allocate ip succeed. ip addr: %v", ip)
	return nil
}

func ReleaseIP(networkName string) error {
	return nil
}

func Connect(networkName string, containerInfo *container.ContainerInfo) error {
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no such network: %s", networkName)
	}

	// 分配容器IP地址
	ip, err := ipAllocator.Allocate(network.IpRange)
	if err != nil {
		return err
	}

	// 创建网络端点
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", containerInfo.Id, networkName),
		IPAddress:   ip,
		Network:     network,
		PortMapping: containerInfo.PortMapping,
	}

	// 调用网络驱动挂载和配置网络端点
	if err = drivers[network.Driver].Connect(network, ep); err != nil {
		return err
	}

	// 给容器的namespace配置容器网络设备IP地址
	if err = configEndpointIpAddressAndRoute(ep, containerInfo); err != nil {
		return err
	}

	// 配置端口映射
	err = configPortMapping(ep, containerInfo)
	if err != nil {
		logrus.Errorf("config port mapping, err: %v", err)
		return err
	}

	return nil
}

// ListNetwork 遍历网络
func ListNetwork() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, _ = fmt.Fprint(w, "Name\tIpRange\tDriver\n")
	for _, nw := range networks {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n",
			nw.Name,
			nw.IpRange.String(),
			nw.Driver)
	}

	if err := w.Flush(); err != nil {
		logrus.Errorf("Flush error %v", err)
		return
	}
}

// DeleteNetwork 删除网络
func DeleteNetwork(networkName string) error {
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no such network: %s", networkName)
	}

	if err := ipAllocator.Release(nw.IpRange, &nw.IpRange.IP); err != nil {
		return fmt.Errorf("remove network gateway ip, err: %v", err)
	}

	if err := drivers[nw.Driver].Delete(*nw); err != nil {
		return fmt.Errorf("remove network driver, err: %v", err)
	}

	return nw.remove(common.DefaultNetworkPath)
}

func configEndpointIpAddressAndRoute(ep *Endpoint, cInfo *container.ContainerInfo) error {
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		logrus.Errorf("fail config endpoint: %v", err)
		return err
	}
	defer enterContainerNetns(&peerLink, cInfo)()

	interfaceIP := *ep.Network.IpRange
	interfaceIP.IP = ep.IPAddress

	if err = setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return fmt.Errorf("%v,%s", ep.Network, err)
	}

	if err = setInterfaceUP(ep.Device.PeerName); err != nil {
		return err
	}

	if err = setInterfaceUP("lo"); err != nil {
		return err
	}

	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")

	logrus.Infof("connect network, network:%v gw:%s", ep.Network, ep.Network.IpRange.IP)

	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IpRange.IP,
		Dst:       cidr,
	}

	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return err
	}

	return nil
}

func enterContainerNetns(enLink *netlink.Link, cInfo *container.ContainerInfo) func() {
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cInfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		logrus.Errorf("error get cotainer net namespace, err: %v", err)
	}

	nsFD := f.Fd()
	runtime.LockOSThread()

	// 修改veth peer 另外一端移到容器的namespace中
	if err = netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		logrus.Errorf("set link netns, err: %v", err)
	}

	// 获取当前的网络namespace
	origns, err := netns.Get()
	if err != nil {
		logrus.Errorf("get current netns, err: %v", err)
	}

	// 设置当前进程到新的网络namespace，并在函数执行完成之后再回复到之前的namespace
	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		logrus.Errorf("error set netns, err: %v", err)
	}

	return func() {
		netns.Set(origns)
		origns.Close()
		runtime.UnlockOSThread()
		f.Close()
	}
}

// 配置端口映射关系
func configPortMapping(ep *Endpoint, cInfo *container.ContainerInfo) error {
	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			logrus.Errorf("port mapping format error, %v", pm)
			continue
		}

		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])

		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		output, err := cmd.Output()
		if err != nil {
			logrus.Errorf("iptables output %v, err: %v", output, err)
			continue
		}
	}

	return nil
}

func (n *Network) load(dumpPath string) error {
	nwConfigFile, err := os.Open(dumpPath)
	if err != nil {
		return err
	}
	defer nwConfigFile.Close()

	nwJson := make([]byte, 2000)
	length, err := nwConfigFile.Read(nwJson)
	if err != nil {
		return err
	}

	err = json.Unmarshal(nwJson[:length], n)
	if err != nil {
		logrus.Errorf("json unmarshal nw info, err: %v", err)
		return err
	}

	n.IpRange = toIPv4(n.IpRange)

	return nil
}

func toIPv4(ipRange *net.IPNet) *net.IPNet {
	_, ipNet, _ := net.ParseCIDR(ipRange.String())
	return ipNet
}

func (n *Network) dump(dumpPath string) error {
	if err := common.Mkdir(dumpPath); err != nil {
		return err
	}

	networkPath := path.Join(dumpPath, n.Name)
	networkFile, err := os.OpenFile(networkPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logrus.Errorf("error: %v", err)
		return err
	}
	defer networkFile.Close()

	networkJson, _ := json.Marshal(n)
	_, err = networkFile.Write(networkJson)
	if err != nil {
		logrus.Errorf("write network file, error: %v", err)
		return err
	}
	return nil
}

func (n *Network) remove(dumpPath string) error {
	fileName := path.Join(dumpPath, n.Name)
	if common.IsNotExist(fileName) {
		return nil
	}

	return os.Remove(fileName)
}
