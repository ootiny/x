package x

import "testing"

func TestCommand(t *testing.T) {
	command := "wg genkey | tee privatekey | wg pubkey > publickey"
	result, err := Command(command)
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	t.Logf("Command result: %s", result)
}
