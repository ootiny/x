package x

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type commandOption struct {
	Stdout       io.Writer
	Stderr       io.Writer
	Sudo         bool
	SudoPassword string
}

func runCommand(command string, option *commandOption) (string, error) {
	// 将command按空格分割成命令和参数
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("command cannot be empty")
	}

	if option == nil {
		return "", fmt.Errorf("option is nil")
	}

	cmd := (*exec.Cmd)(nil)
	if option.Sudo {
		sudoArgs := append([]string{"-S"}, parts...)
		cmd = exec.Command("sudo", sudoArgs...)
	} else {
		cmd = exec.Command(parts[0], parts[1:]...)
	}

	// 获取 StdinPipe 用于写入密码
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("error creating stdin pipe for sudo: %v", err)
	}
	// 获取 StdoutPipe 和 StderrPipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("error creating stdout pipe for sudo: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("error creating stderr pipe for sudo: %v", err)
	}

	var outputBuf bytes.Buffer

	// 启动命令
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("error starting sudo command: %v", err)
	}

	results := make(chan error)

	go func() {
		defer stdin.Close()
		if option.Sudo && option.SudoPassword != "" {
			_, err := io.WriteString(stdin, option.SudoPassword+"\n") // 发送密码并加换行符
			results <- err
		} else {
			results <- nil
		}
	}()

	go func() {
		defer stdout.Close()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if option.Stdout != nil {
				fmt.Fprintln(option.Stdout, line)
			}
			outputBuf.WriteString(line + "\n")
		}
		results <- scanner.Err()
	}()

	go func() {
		defer stderr.Close()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if option.Stderr != nil {
				fmt.Fprintln(option.Stderr, line)
			}
			outputBuf.WriteString(line + "\n")
		}
		results <- scanner.Err()
	}()

	for i := 0; i < 3; i++ {
		result := <-results
		if result != nil {
			return outputBuf.String(), result
		}
	}

	return outputBuf.String(), cmd.Wait()
}

func XCommand(command string) (string, error) {
	return runCommand(command, &commandOption{
		Stdout:       os.Stdout,
		Stderr:       os.Stderr,
		Sudo:         false,
		SudoPassword: "",
	})
}

func XSudoCommand(command string, password string) (string, error) {
	return runCommand(command, &commandOption{
		Stdout:       os.Stdout,
		Stderr:       os.Stderr,
		Sudo:         true,
		SudoPassword: password,
	})
}
