package main

import (
	"os"
	"strings"
	"time"

	"github.com/ootiny/x"
)

func main() {
	// privateKey, err := x.NewCommand().Eval("wg genkey")
	// if err != nil {
	// 	x.LogErrorf("aCommand failed : %v", err)
	// }
	// privateKey = strings.TrimSpace(privateKey)

	// x.LogInfof("Private key: %s\n", privateKey)

	publicKey, err := x.NewCommand(&x.CommandConfig{
		Stdin:  nil,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}).Eval(`echo 2BZkqdp3FikyjANM+XUy8aNfG/uZjpuXISzItGe9QG0= | wg pubkey`)
	if err != nil {
		x.LogErrorf("aCommand failed : %v", err)
	}
	publicKey = strings.TrimSpace(publicKey)

	x.LogInfof("Public key: %s\n", publicKey)

	time.Sleep(time.Second)
}
