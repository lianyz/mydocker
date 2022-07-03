/*
@Time : 2022/7/3 17:03
@Author : lianyz
@Description :
*/

package network

import (
	"net"
	"testing"
)

func testAllocate(t *testing.T) {
	_, ipnet, _ := net.ParseCIDR("192.168.0.1/24")
	ip, _ := ipAllocator.Allocate(ipnet)
	t.Logf("alloc ip: %v", ip)
}

func TestRelase(t *testing.T) {
	ip, ipnet, _ := net.ParseCIDR("192.168.0.1/24")
	ipAllocator.Release(ipnet, &ip)
}
