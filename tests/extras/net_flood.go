package extras

import (
	"net"
	"time"
)

func NetProbe(addr string, attempts int) int {
	if addr == "" {
		addr = "1.1.1.1:53"
	}
	if attempts <= 0 {
		attempts = 20
	}
	fails := 0
	for i := 0; i < attempts; i++ {
		c, err := net.DialTimeout("udp", addr, 250*time.Millisecond)
		if err != nil {
			fails++
			continue
		}
		_ = c.Close()
	}
	return fails
}
