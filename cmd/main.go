package main

import (
	"fmt"
	"strings"

	"github.com/ootiny/x"
)

func RestoreEsxiVM(vmid int, snapshotId int) error {
	opt := x.NewSSHOptionWithPassword("root", "192.168.1.11", "Tscc0805@")
	opt.Expect = func(output string) (string, error) {
		fmt.Println(output)
		if strings.Contains(output, "Tscc0805@") {
			return "Tscc0805@\n", nil
		}
		return "", nil
	}
	if _, err := x.SSH(fmt.Sprintf("vim-cmd vmsvc/snapshot.revert %d %d 0", vmid, snapshotId), opt); err != nil {
		return err
	} else {
		return nil
	}
}

func main() {
	fmt.Println(RestoreEsxiVM(56, 1))
}
