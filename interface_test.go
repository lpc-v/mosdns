package main

import (
	"fmt"
	"net"
	"testing"
)
func TestAddrs(t *testing.T) {
	itf, _ := net.InterfaceByName("en0")
	addrs, _ := itf.Addrs()
	// fmt.Println(len(addrs))
	for _, addr := range addrs {
		fmt.Println(addr.Network(),addr.String())
	}
}