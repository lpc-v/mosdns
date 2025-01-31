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

package server

import (
	"context"
	"fmt"
	"github.com/IrineSistiana/mosdns/v4/pkg/pool"
	"github.com/IrineSistiana/mosdns/v4/pkg/query_context"
	"github.com/IrineSistiana/mosdns/v4/pkg/utils"
	"github.com/miekg/dns"
	"go.uber.org/zap"
	"io"
	"net"
)

type udpResponseWriter struct {
	c       net.PacketConn
	to      net.Addr
	udpSize int
}

func (w *udpResponseWriter) Write(m *dns.Msg) error {
	m.Truncate(w.udpSize)
	b, buf, err := pool.PackBuffer(m)
	if err != nil {
		return err
	}
	defer buf.Release()
	_, err = w.c.WriteTo(b, w.to)
	return err
}

func (s *Server) ServeUDP(c net.PacketConn) error {
	defer c.Close()

	handler := s.opts.DNSHandler
	if handler == nil {
		return errMissingDNSHandler
	}

	closer := io.Closer(c)
	if ok := s.trackCloser(&closer, true); !ok {
		return ErrServerClosed
	}
	defer s.trackCloser(&closer, false)

	listenerCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	readBuf := pool.GetBuf(64 * 1024)
	defer readBuf.Release()
	rb := readBuf.Bytes()

	for {
		n, from, err := c.ReadFrom(rb)
		if err != nil {
			if s.Closed() {
				return ErrServerClosed
			}
			return fmt.Errorf("unexpected read err: %w", err)
		}

		req := new(dns.Msg)
		if err := req.Unpack(rb[:n]); err != nil {
			s.opts.Logger.Warn("invalid msg", zap.Error(err), zap.Binary("msg", rb[:n]))
			continue
		}

		go func() {
			meta := new(query_context.RequestMeta)
			meta.FromUDP = true
			if clientIP := utils.GetIPFromAddr(from); clientIP != nil {
				meta.ClientIP = clientIP
			} else {
				s.opts.Logger.Warn("failed to acquire client ip addr")
			}

			w := &udpResponseWriter{c: c, to: from, udpSize: getUDPSize(req)}

			if err := handler.ServeDNS(listenerCtx, req, w, meta); err != nil {
				s.opts.Logger.Warn("handler err", zap.Error(err))
			}
		}()
	}
}

func getUDPSize(m *dns.Msg) int {
	var s uint16
	if opt := m.IsEdns0(); opt != nil {
		s = opt.UDPSize()
	}
	if s < dns.MinMsgSize {
		s = dns.MinMsgSize
	}
	return int(s)
}
