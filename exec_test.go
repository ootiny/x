package x

import "testing"

func TestCommand(t *testing.T) {
	if _, err := NewCommand().Eval(`ps -a | grep bin | grep b`); err != nil {
		t.Fatalf("Command failed : %v", err)
	}
}
