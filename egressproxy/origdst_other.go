//go:build !linux

package main

import (
	"errors"
	"net"
)

func originalDestination(_ net.Conn) (string, net.IP, error) {
	return "", nil, errors.New("SO_ORIGINAL_DST is only supported on linux")
}
