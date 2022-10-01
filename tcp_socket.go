package gio

import (
	"errors"
	"golang.org/x/sys/unix"
	"net"
	"syscall"
)

func GetTCPSockAddr(proto, addr string) (sa unix.Sockaddr, family int, tcpAddr *net.TCPAddr, ipv6only bool, err error) {
	var tcpVersion string
	tcpAddr, err = net.ResolveTCPAddr(proto, addr)
	if err != nil {
		return
	}
	tcpVersion, err = determineTCPProto(proto, tcpAddr)
	if err != nil {
		return
	}
	switch tcpVersion {
	case "tcp4":
		family = unix.AF_INET
		sa, err = ipToSockAddr(family, tcpAddr.IP, tcpAddr.Port, "")
	case "tcp6":
		ipv6only = true
		fallthrough
	case "tcp":
		family = unix.AF_INET6
		sa, err = ipToSockAddr(family, tcpAddr.IP, tcpAddr.Port, tcpAddr.Zone)
	default:
		err = errors.New("")
	}
	return
}

func determineTCPProto(proto string, addr *net.TCPAddr) (string, error) {
	if addr.IP.To4() != nil {
		return "tcp4", nil
	}
	if addr.IP.To16() != nil {
		return "tcp6", nil
	}
	switch proto {
	case "tcp", "tcp4", "tcp6":
		return proto, nil
	}
	return "", errors.New("")
}

func ipToSockAddr(family int, ip net.IP, port int, zone string) (unix.Sockaddr, error) {
	switch family {
	case syscall.AF_INET:
		sa, err := ipToSockAddrInet4(ip, port)
		if err != nil {
			return nil, err
		}
		return &sa, nil
	case syscall.AF_INET6:
		sa, err := ipToSockAddrInet6(ip, port, zone)
		if err != nil {
			return nil, err
		}
		return &sa, nil
	}
	return nil, &net.AddrError{Err: "invalid address family", Addr: ip.String()}
}

func ipToSockAddrInet4(ip net.IP, port int) (unix.SockaddrInet4, error) {
	if len(ip) == 0 {
		ip = net.IPv4zero
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return unix.SockaddrInet4{}, &net.AddrError{Err: "non-IPv4 address", Addr: ip.String()}
	}
	sa := unix.SockaddrInet4{Port: port}
	copy(sa.Addr[:], ip4)
	return sa, nil
}

func ipToSockAddrInet6(ip net.IP, port int, zone string) (unix.SockaddrInet6, error) {
	if len(ip) == 0 || ip.Equal(net.IPv4zero) {
		ip = net.IPv6zero
	}
	ip6 := ip.To16()
	if ip6 == nil {
		return unix.SockaddrInet6{}, &net.AddrError{Err: "non-IPv6 address", Addr: ip.String()}
	}

	sa := unix.SockaddrInet6{Port: port}
	copy(sa.Addr[:], ip6)
	iface, err := net.InterfaceByName(zone)
	if err != nil {
		return sa, nil
	}
	sa.ZoneId = uint32(iface.Index)

	return sa, nil
}
