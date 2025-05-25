package main

import (
	"strings"

	"github.com/ootiny/x"
)

func main() {
	opt := x.NewSSHOptionWithPassword("tianshuo", "192.168.1.81", "World2019")
	opt.Expect = func(output string) (string, error) {
		if strings.Contains(output, "[sudo] password for") {
			return "World2019\n", nil
		}

		return "", nil
	}

	// if _, err := x.SSH("sudo -S apt-get update", opt); err != nil {
	// 	panic(err)
	// }

	if err := x.SCP("/Users/tianshuo/Downloads/debian-12.11.0-arm64-DVD-1.iso", "/home/tianshuo/1.txt", opt); err != nil {
		panic(err)
	}
}
