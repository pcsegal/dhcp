package server4

import (
	"fmt"
	"net"
	"os"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"golang.org/x/sys/unix"
)

// NewIPv4UDPConn returns a UDP connection bound to both the interface and port
// given based on a IPv4 DGRAM socket. The UDP connection allows broadcasting.
//
// The interface must already be configured.
func NewIPv4UDPConn(iface string, port int) (net.PacketConn, error) {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_UDP)
	if err != nil {
		return nil, fmt.Errorf("cannot get a UDP socket: %v", err)
	}
	f := os.NewFile(uintptr(fd), "")
	// net.FilePacketConn dups the FD, so we have to close this in any case.
	defer f.Close()

	// Allow broadcasting.
	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_BROADCAST, 1); err != nil {
		return nil, fmt.Errorf("cannot set broadcasting on socket: %v", err)
	}
	// Allow reusing the addr to aid debugging.
	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
		return nil, fmt.Errorf("cannot set reuseaddr on socket: %v", err)
	}
	if len(iface) != 0 {
		// Bind directly to the interface.
		if err := dhcpv4.BindToInterface(fd, iface); err != nil {
			return nil, fmt.Errorf("cannot bind to interface %s: %v", iface, err)
		}
	}
	// Bind to the port.
	if err := unix.Bind(fd, &unix.SockaddrInet4{Port: port}); err != nil {
		return nil, fmt.Errorf("cannot bind to port %d: %v", port, err)
	}

	return net.FilePacketConn(f)
}
