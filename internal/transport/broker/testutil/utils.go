// Package testutil provides helpful functions for testing the MQTT broker
package testutil

import (
	"fmt"
	"net"
	"time"
)

// WaitForPort waits until a TCP address is recieving connections.
func WaitForPort(address string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn, err := net.Dial("tcp", "127.0.0.1"+address)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("port %s not ready after %s", address, timeout)
}
