/*
@Time : 2022/6/29 23:03
@Author : lianyz
@Description :
*/

package network

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"net"
	"strings"
	"time"
)

// BridgeNetworkDriver 桥接驱动
type BridgeNetworkDriver struct {
}

func (d *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (d *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	ip, ipRange, _ := net.ParseCIDR(subnet)
	ipRange.IP = ip
	n := &Network{
		Name:    name,
		IpRange: ipRange,
		Driver:  d.Name(),
	}
	err := d.initBridge(n)
	if err != nil {
		logrus.Errorf("error init bridge: %v", err)
		return nil, err
	}

	return n, err
}

func (d *BridgeNetworkDriver) Delete(network Network) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	return netlink.LinkDel(br)
}

func (d *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	la := netlink.NewLinkAttrs()
	la.Name = endpoint.ID[:5]
	la.MasterIndex = br.Attrs().Index

	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + endpoint.ID[:5],
	}

	if err = netlink.LinkAdd(&endpoint.Device); err != nil {
		logrus.Errorf("add endpoint device, err: %v", err)
		return err
	}

	if err = netlink.LinkSetUp(&endpoint.Device); err != nil {
		logrus.Errorf("setup endpoint device: %v", err)
		return err
	}

	return nil
}

func (d *BridgeNetworkDriver) Disconnect(network Network, endpoint *Endpoint) error {
	return nil
}

func (d *BridgeNetworkDriver) initBridge(n *Network) error {
	bridgeName := n.Name
	if err := createBridgeInterface(bridgeName); err != nil {
		logrus.Errorf("add bridge: %s, err: %v", bridgeName, err)
		return err
	}

	logrus.Infof("init bridge 0 ipNet:%v gatewayIP:%v", n.IpRange, n.IpRange.IP)

	gatewayIP := *n.IpRange
	gatewayIP.IP = n.IpRange.IP

	logrus.Infof("init bridge 1 ipNet:%v gatewayIP:%v", n.IpRange, n.IpRange.IP)

	if err := setInterfaceIP(bridgeName, gatewayIP.String()); err != nil {
		logrus.Errorf("assigning address: %s on bridge: %s with an error: %v",
			gatewayIP, bridgeName, err)
		return err
	}

	if err := setInterfaceUP(bridgeName); err != nil {
		logrus.Errorf("set bridge up: %s, err: %v", bridgeName, err)
		return err
	}

	if err := setSNAT(bridgeName, n.IpRange); err != nil {
		logrus.Errorf("setting iptables for %s, err: %v", bridgeName, err)
		return err
	}

	if err := setBridgeForward(bridgeName); err != nil {
		logrus.Errorf("setting iptables for %s, err: %v", bridgeName, err)
		return err
	}

	return nil
}

func createBridgeInterface(bridgeName string) error {
	_, err := net.InterfaceByName(bridgeName)
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}

	linkAttr := netlink.NewLinkAttrs()
	linkAttr.Name = bridgeName

	br := &netlink.Bridge{LinkAttrs: linkAttr}
	if err := netlink.LinkAdd(br); err != nil {
		logrus.Errorf("bridge creation failed for bridge %s, err: %v", bridgeName, err)
		return err
	}

	return nil
}

func setInterfaceIP(name string, rawIP string) error {
	retries := 2
	var link netlink.Link
	var err error
	for i := 0; i < retries; i++ {
		link, err = netlink.LinkByName(name)
		if err == nil {
			break
		}
		logrus.Debugf("error retrieving ew bridge netlink link [ %s ]...retrying", name)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		logrus.Errorf("abandoning retrieving the new bridge link from netlink, "+
			"Run [ ip link ] to troubleshoot the error: %v", err)
		return err
	}
	ipNet, err := netlink.ParseIPNet(rawIP)
	if err != nil {
		return err
	}

	addr := &netlink.Addr{
		IPNet: ipNet,
		Label: "",
		Flags: 0,
		Scope: 0,
	}

	return netlink.AddrAdd(link, addr)
}

func setInterfaceUP(interfaceName string) error {
	link, err := netlink.LinkByName(interfaceName)
	if err != nil {
		logrus.Errorf("retrieving a link, err: %v", err)
		return err
	}

	if err := netlink.LinkSetUp(link); err != nil {
		logrus.Errorf("enabling interface for %s, err: %v", interfaceName, err)
		return err
	}

	return nil
}

func setSNAT(bridgeName string, subnet *net.IPNet) error {
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s -j MASQUERADE",
		subnet.String())

	return setupIPTables(iptablesCmd)
}

func setBridgeForward(bridgeName string) error {
	iptablesCmd := fmt.Sprintf("-A FORWARD -i %s -j ACCEPT", bridgeName)

	return setupIPTables(iptablesCmd)
}
