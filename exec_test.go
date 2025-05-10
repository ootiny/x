package x

import "testing"

func TestCommand(t *testing.T) {
	command := "ls -la /"
	_, err := Command(command)
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
}

func TestSudoCommand(t *testing.T) {
	command := "ls -la /"
	_, err := SudoCommand(command, "World2019")
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
}
