package store

// This file exists on this branch only. It is the near miss for the third: an
// outbound connection opened where it is needed rather than taken as a field,
// which is the dependency that starts calling somewhere and is found by a
// firewall months later.

import "net"

// Reachable reports whether the address answers. It opens a socket, so a suite
// carrying it is a suite that needs a network.
func Reachable(address string) bool {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return false
	}
	if closeErr := conn.Close(); closeErr != nil {
		return false
	}
	return true
}
