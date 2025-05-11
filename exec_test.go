package x

import (
	"testing"
)

func TestCommand(t *testing.T) {
	if _, err := NewCommand().Eval("az group list"); err != nil {
		t.Fatalf("aCommand failed : %v", err)
	}
}
