package main

import "github.com/ootiny/x"

func main() {
	remote := x.NewSSHClient("tianshuo", "192.168.1.7", "World2019")

	if err := remote.Open(); err != nil {
		panic(err)
	}

	defer remote.Close()

	remote.SSH("ls -l")
}
