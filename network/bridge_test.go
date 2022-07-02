/*
@Time : 2022/7/2 08:25
@Author : lianyz
@Description :
*/

package network

import (
	"github.com/lianyz/mydocker/container"
	"testing"
)

func TestBridgeInit(t *testing.T) {
	d := BridgeNetworkDriver{}
	_, err := d.Create("192.168.0.1/24", "test-bridge")
	t.Logf("err: %v", err)
}

func TestBridgeConnect(t *testing.T) {
	ep := Endpoint{
		ID: "test container",
	}

	n := Network{
		Name: "test-bridge",
	}

	d := BridgeNetworkDriver{}
	err := d.Connect(&n, &ep)
	t.Logf("err: %v", err)
}

func TestNetworkConnect(t *testing.T) {
	cInfo := &container.ContainerInfo{
		Id:  "test-container",
		Pid: "15438",
	}

	d := BridgeNetworkDriver{}
	n, err := d.Create("192.168.0.1/24", "test-bridge")
	t.Logf("err: %v", n)

	networks[n.Name] = n
	err = Connect(n.Name, cInfo)
	t.Logf("err: %v", err)
}

func TestLoad(t *testing.T) {
	n := Network{
		Name: "test-bridge",
	}

	n.load("/var/run/mydocker/network/network/test-bridge")
	t.Logf("network: %v", n)
}


