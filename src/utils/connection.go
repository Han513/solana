// utils/connection.go
package utils

import (
	"fmt"
	"net"
	"time"
)

// TCPConnectionChecker provides TCP connection testing functionality
type TCPConnectionChecker struct {
	timeout time.Duration
}

func NewTCPConnectionChecker(timeout time.Duration) *TCPConnectionChecker {
	return &TCPConnectionChecker{
		timeout: timeout,
	}
}

func (c *TCPConnectionChecker) TestConnection(host string, port string) error {
	address := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.DialTimeout("tcp", address, c.timeout)
	if err != nil {
		return fmt.Errorf("TCP connection to %s failed: %v", address, err)
	}
	defer conn.Close()
	return nil
}
