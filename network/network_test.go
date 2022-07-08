/*
@Time : 2022/7/7 12:17
@Author : lianyz
@Description :
*/

package network

import (
	"testing"
)

func TestCreateNetwork(t *testing.T) {
	err := CreateNetwork("bridge", "192.168.1.2/24", "testnet")
	if err != nil {
		t.Errorf("create network failed. err: %v", err)
	}
}
