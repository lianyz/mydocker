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
	"os/exec"
	"strings"
	"time"
)

// BridgeNetworkDriver 桥接驱动
type BridgeNetworkDriver struct {
}

func (d *BridgeNetworkDriver) Name() string {

	// todo

	return "bridge"
}

func (d *BridgeNetworkDriver) Delete(network Network) error {

	// todo

	return nil
}

func (d *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {

	// todo

	return nil
}

func (d *BridgeNetworkDriver) Disconnect(network Network, endpoint *Endpoint) error {

	// todo

	return nil
}

func (d *BridgeNetworkDriver) Create(subnet, name string) (*Network, error) {
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

func (d *BridgeNetworkDriver) initBridge(n *Network) error {
	bridgeName := n.Name
	if err := createBridgeInterface(bridgeName); err != nil {
		logrus.Errorf("add bridge: %s, err: %v", bridgeName, err)
		return err
	}

	gatewayIP := *n.IpRange
	gatewayIP.IP = n.IpRange.IP

	if err := setInterfaceIP(bridgeName, gatewayIP.String()); err != nil {
		logrus.Errorf("assigning address: %s on bridge: %s with an error: %v",
			gatewayIP, bridgeName, err)
		return err
	}

	if err := setInterfaceUP(bridgeName); err != nil {
		logrus.Errorf("set bridge up: %s, err: %v", bridgeName, err)
		return err
	}

	if err := setupIPTables(bridgeName, n.IpRange); err != nil {
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
		IPNet:     ipNet,
		Peer:      ipNet,
		Label:     "",
		Flags:     0,
		Scope:     0,
		Broadcast: nil,
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

func setupIPTables(bridgeName string, subnet *net.IPNet) error {
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE",
		subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	output, err := cmd.Output()
	if err != nil {
		logrus.Errorf("iptables output: %v, err: %v", output, err)
		return err
	}
	return nil
}