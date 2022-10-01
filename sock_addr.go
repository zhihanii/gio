package gio

import (
	"golang.org/x/sys/unix"
	"net"
)

var ipv4InIPv6Prefix = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff}

func SockAddrToTCPOrUnixAddr(sa unix.Sockaddr) net.Addr {
	switch sa := sa.(type) {
	case *unix.SockaddrInet4:
		ip := sockAddrInet4ToIP(sa)
		return &net.TCPAddr{IP: ip, Port: sa.Port}
	}
	return nil
}

func sockAddrInet4ToIP(sa *unix.SockaddrInet4) net.IP {
	ip := make([]byte, 16)
	copy(ip[0:12], ipv4InIPv6Prefix)
	copy(ip[12:16], sa.Addr[:])
	return ip
}
