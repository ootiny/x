package main

import "github.com/ootiny/x"

func main() {
	remote := x.NewSSHClient(x.SSHConfig{
		User:     "tianshuo",
		Host:     "192.168.1.7",
		Password: "World2019",
	})

	if err := remote.Open(); err != nil {
		panic(err)
	}

	defer remote.Close()

	remote.SSH("ls -l")
}
