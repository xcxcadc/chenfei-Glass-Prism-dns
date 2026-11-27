package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

func warpSocksReachable(ctx context.Context, socksAddr string) bool {
	socksAddr = strings.TrimSpace(socksAddr)
	if socksAddr == "" {
		return false
	}
	probeCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	conn, err := (&net.Dialer{Timeout: 500 * time.Millisecond}).DialContext(probeCtx, "tcp", socksAddr)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// dialSOCKS5 establishes a TCP tunnel through a no-auth RFC 1928 SOCKS5 proxy
// (such as local wireproxy / warp-cli on 127.0.0.1:40000).
func dialSOCKS5(ctx context.Context, proxyAddr, targetHost, targetPort string) (net.Conn, error) {
	portNum, err := strconv.Atoi(targetPort)
	if err != nil || portNum < 1 || portNum > 65535 {
		return nil, fmt.Errorf("invalid target port %q", targetPort)
	}
	dialer := &net.Dialer{Timeout: 2 * time.Second, KeepAlive: 30 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", proxyAddr)
	if err != nil {
		return nil, err
	}
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-done:
		}
	}()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	}

	// Greeting: VER=5, NMETHODS=1, METHOD=0 (no auth)
	if _, err := conn.Write([]byte{0x05, 0x01, 0x00}); err != nil {
		_ = conn.Close()
		return nil, err
	}
	methodResp := make([]byte, 2)
	if _, err := io.ReadFull(conn, methodResp); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if methodResp[0] != 0x05 || methodResp[1] != 0x00 {
		_ = conn.Close()
		return nil, fmt.Errorf("socks5 auth rejected: %x %x", methodResp[0], methodResp[1])
	}

	// Connect request: VER=5, CMD=1 (CONNECT), RSV=0
	req := make([]byte, 0, 264)
	req = append(req, 0x05, 0x01, 0x00)
	if ip := net.ParseIP(targetHost); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			req = append(req, 0x01)
			req = append(req, ip4...)
		} else {
			req = append(req, 0x04)
			req = append(req, ip.To16()...)
		}
	} else {
		if len(targetHost) == 0 || len(targetHost) > 255 {
			_ = conn.Close()
			return nil, fmt.Errorf("invalid socks5 target host %q", targetHost)
		}
		req = append(req, 0x03, byte(len(targetHost)))
		req = append(req, targetHost...)
	}
	var portBuf [2]byte
	binary.BigEndian.PutUint16(portBuf[:], uint16(portNum))
	req = append(req, portBuf[:]...)

	if _, err := conn.Write(req); err != nil {
		_ = conn.Close()
		return nil, err
	}

	replyHead := make([]byte, 4)
	if _, err := io.ReadFull(conn, replyHead); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if replyHead[0] != 0x05 || replyHead[1] != 0x00 {
		_ = conn.Close()
		return nil, fmt.Errorf("socks5 connect failed: rep=0x%02x", replyHead[1])
	}

	// Drain bound address + port
	var discardLen int
	switch replyHead[3] {
	case 0x01:
		discardLen = 4 + 2
	case 0x04:
		discardLen = 16 + 2
	case 0x03:
		lenByte := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenByte); err != nil {
			_ = conn.Close()
			return nil, err
		}
		discardLen = int(lenByte[0]) + 2
	default:
		_ = conn.Close()
		return nil, errors.New("socks5 invalid address type in reply")
	}
	if _, err := io.CopyN(io.Discard, conn, int64(discardLen)); err != nil {
		_ = conn.Close()
		return nil, err
	}

	_ = conn.SetDeadline(time.Time{})
	return conn, nil
}

// dialViaWARP dials target host:port through the local WARP SOCKS5 proxy.
// family may be "ipv6", "ipv4", or "" (let proxy resolve hostname).
func dialViaWARP(ctx context.Context, socksAddr, host, port, family string) (net.Conn, error) {
	if strings.TrimSpace(socksAddr) == "" {
		return nil, errors.New("warp socks address is empty")
	}
	if family == "ipv6" || family == "ipv4" {
		wantV6 := family == "ipv6"
		if addrs, err := lookupPublicIPs(ctx, host); err == nil {
			var candidates []string
			for _, a := range addrs {
				if !isUsableDirectIP(a.IP) {
					continue
				}
				if (a.IP.To4() == nil) == wantV6 {
					candidates = append(candidates, a.IP.String())
				}
			}
			for _, ipStr := range candidates {
				if conn, dialErr := dialSOCKS5(ctx, socksAddr, ipStr, port); dialErr == nil {
					return conn, nil
				}
			}
		}
	}
	// Fallback to domain-level SOCKS5 connect (works when domain is IPv4-only or DNS is inside WARP)
	return dialSOCKS5(ctx, socksAddr, host, port)
}
