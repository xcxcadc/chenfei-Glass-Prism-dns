//go:build linux

package main

import (
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"syscall"
	"unsafe"
)

const (
	soOriginalDst   = 80
	ip6tSoOrigDst   = 80
	sockaddrStorage = 128
)

// originalDestination retrieves the pre-NAT destination IP:port from a Linux
// netfilter REDIRECT/DNAT socket using SO_ORIGINAL_DST / IP6T_SO_ORIGINAL_DST.
func originalDestination(conn net.Conn) (string, net.IP, error) {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return "", nil, errors.New("not a TCP connection")
	}
	rawConn, err := tcpConn.SyscallConn()
	if err != nil {
		return "", nil, err
	}

	var addr string
	var dstIP net.IP
	var sockErr error

	ctrlErr := rawConn.Control(func(fd uintptr) {
		var buf [sockaddrStorage]byte
		size := uint32(len(buf))

		// Try IPv4 SO_ORIGINAL_DST (SOL_IP = 0, 80)
		_, _, errno := syscall.Syscall6(
			syscall.SYS_GETSOCKOPT,
			fd,
			uintptr(syscall.SOL_IP),
			uintptr(soOriginalDst),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)),
			0,
		)
		if errno == 0 && size >= 16 {
			family := *(*uint16)(unsafe.Pointer(&buf[0]))
			if family == syscall.AF_INET {
				port := binary.BigEndian.Uint16(buf[2:4])
				ip := net.IPv4(buf[4], buf[5], buf[6], buf[7])
				dstIP = ip
				addr = net.JoinHostPort(ip.String(), strconv.Itoa(int(port)))
				return
			}
		}

		// Try IPv6 IP6T_SO_ORIGINAL_DST (SOL_IPV6 = 41, 80)
		size = uint32(len(buf))
		_, _, errno = syscall.Syscall6(
			syscall.SYS_GETSOCKOPT,
			fd,
			uintptr(syscall.SOL_IPV6),
			uintptr(ip6tSoOrigDst),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)),
			0,
		)
		if errno == 0 && size >= 28 {
			family := *(*uint16)(unsafe.Pointer(&buf[0]))
			if family == syscall.AF_INET6 {
				port := binary.BigEndian.Uint16(buf[2:4])
				ip := make(net.IP, net.IPv6len)
				copy(ip, buf[8:24])
				if v4 := ip.To4(); v4 != nil {
					ip = v4
				}
				dstIP = ip
				addr = net.JoinHostPort(ip.String(), strconv.Itoa(int(port)))
				return
			}
		}
		sockErr = errors.New("SO_ORIGINAL_DST unavailable")
	})
	if ctrlErr != nil {
		return "", nil, ctrlErr
	}
	if sockErr != nil {
		return "", nil, sockErr
	}
	return addr, dstIP, nil
}
