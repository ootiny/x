package main

import (
	"strings"

	"github.com/ootiny/x"
)

func RestoreEsxiVM(vmid int, snapshotId int) error {
	remote := x.NewSSHClient("root", "192.168.1.11", "Tscc0805@")
	remote.AuthKeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
		answers = make([]string, len(questions))
		for i := range questions {
			answers[i] = "Tscc0805@"
		}
		return answers, nil
	})
	if err := remote.Open(); err != nil {
		return err
	}
	defer remote.Close()

	if _, err := remote.SSH("vim-cmd vmsvc/snapshot.revert %d %d 0", vmid, snapshotId); err != nil {
		return err
	}
	return nil
}

func main() {
	remote := x.NewSSHClient("tianshuo", "192.168.1.7", "World2019")
	remote.SetExpect(func(output string) (string, error) {
		if strings.Contains(output, "assword") {
			return "World2019\n", nil
		}
		return "", nil
	})
	if err := remote.Open(); err != nil {
		panic(err)
	}
	defer remote.Close()

	remote.SetStdout(nil)
	remote.SetStderr(nil)

	if output, err := remote.SudoSSH("pwdd"); err != nil {
		panic(err)
	} else {
		x.LogInfo(output)
	}

	// if err := remote.SCP("~/Downloads/test.txt", "/tmp/test.txt"); err != nil {
	// 	panic(err)
	// }

	// homeDir, err := remote.RemoteHomeDir()
	// if err != nil {
	// 	panic(err)
	// }

}
