package x

import (
	"testing"
)

func TestCommand(t *testing.T) {
	if _, err := NewCommand().Eval("wg genkey | tee privatekey | wg pubkey > publickey"); err != nil {
		t.Fatalf("aCommand failed : %v", err)
	}
}
