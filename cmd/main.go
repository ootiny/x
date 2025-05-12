package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ootiny/x"
)

func main() {
	if _, err := x.NewCommand(&x.CommandConfig{
		Stdin:  strings.NewReader("World2019\n"),
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}).Eval("sudo ls -la /"); err != nil {
		fmt.Println(err)
	}
}
