package x

import (
	"net"
	"time"
)

func WaitForTCP(network string, ip string, port uint16, timeout time.Duration) bool {
	deadLine := time.Now().Add(timeout)

	for time.Now().Before(deadLine) {
		ColorPrintf("blue", "Waiting for %s %s:%d", network, ip, port)
		conn, err := net.DialTimeout(network, Sprintf("%s:%d", ip, port), time.Second*3)
		if err != nil {
			ColorPrintf("blue", " ...\n")
			time.Sleep(time.Second)
			continue
		} else {
			ColorPrintf("green", " OK\n")
			conn.Close()
			return true
		}
	}

	return false
}
