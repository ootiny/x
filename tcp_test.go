package x

import (
	"testing"
	"time"
)

func TestWaitForTCP(t *testing.T) {
	ok := WaitForTCP("tcp", "192.168.1.7", 22, time.Second*60)
	if !ok {
		t.Fatal("failed to wait for tcp")
	}
}
