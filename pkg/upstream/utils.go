/*
 * Copyright (C) 2020-2022, IrineSistiana
 *
 * This file is part of mosdns.
 *
 * mosdns is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * mosdns is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package upstream

import (
	"context"
	"fmt"
	"net"

	"golang.org/x/net/proxy"
)

func dialTCP(ctx context.Context, addr, socks5 string, dialer *net.Dialer) (net.Conn, error) {
	if len(socks5) > 0 {
		socks5Dialer, err := proxy.SOCKS5("tcp", socks5, nil, dialer)
		if err != nil {
			return nil, fmt.Errorf("failed to init socks5 dialer: %w", err)
		}
		return socks5Dialer.(proxy.ContextDialer).DialContext(ctx, "tcp", addr)
	}

	return dialer.DialContext(ctx, "tcp", addr)
}

func getIPv4FromInterfaceName(name string) net.IP {
	intf, err := net.InterfaceByName(name)
	if err != nil {
		return nil
	}
	addrs, _ := intf.Addrs()
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if ok {
			ipv4 := ipnet.IP.To4()
			if ipv4 != nil {
				return ipv4
			}
		}
	}
	return nil
}

func getUDPAddrFromInterfaceName(name string) *net.UDPAddr {
	ipv4 := getIPv4FromInterfaceName(name)
	if ipv4 == nil {
		return nil
	}
	return &net.UDPAddr{
		IP: ipv4,
		Port: 0,
	}
}
