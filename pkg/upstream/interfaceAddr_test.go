package upstream

import (
	"fmt"
	"net"
	"testing"
)


func TestInterfaceAddr(t *testing.T) {
	addr := getIPv4FromInterfaceName("en0")
	ipNet, _ := addr.(*net.IPNet)
	ip := ipNet.IP
	d := net.Dialer{
		LocalAddr: &net.TCPAddr{
			IP: ip,
			Port: 0,
		},
	}
	conn, err :=  d.Dial("tcp", "49.232.135.32:8080")
	if err != nil {
		fmt.Println(err)
	}
	conn.Write([]byte("hello"))
}