package main

import (
	"strings"

	"github.com/ootiny/x"
)

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

}
