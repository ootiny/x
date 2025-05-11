package x

import "testing"

func TestCommand(t *testing.T) {
	if _, err := NewCommand().Eval(`wg genkey > privatekey`); err != nil {
		t.Fatalf("Command failed : %v", err)
	}

	if _, err := NewCommand().Eval(`wg pubkey < privatekey > publickey`); err != nil {
		t.Fatalf("Command failed : %v", err)
	}
}
