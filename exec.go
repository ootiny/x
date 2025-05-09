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

	shellCommand := []string{"sh", "-c", command}

	if option.Sudo {
		// Prepend sudo -S to the shell command
		sudoArgs := append([]string{"-S"}, shellCommand...)
		cmd = exec.Command("sudo", sudoArgs...)
	} else {
		cmd = exec.Command(shellCommand[0], shellCommand[1:]...)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", Errorf("error creating stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", Errorf("error creating stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", Errorf("error creating stderr pipe: %v", err)
	}

	var outputBuf bytes.Buffer
	errorCHan := make(chan error, 3)

	go func() {
		defer stdin.Close()
		if option.Sudo && option.SudoPassword != "" {
			_, err := io.WriteString(stdin, option.SudoPassword+"\n") // 发送密码并加换行符
			errorCHan <- err
		} else {
			errorCHan <- nil
		}
	}()

	go func() {
		defer stdout.Close()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if option.Stdout != nil {
				_, _ = Fprintln(option.Stdout, line)
			}
			outputBuf.WriteString(line + "\n")
		}
		errorCHan <- scanner.Err()
	}()

	go func() {
		defer stderr.Close()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if option.Stderr != nil {
				_, _ = fmt.Fprintln(option.Stderr, line)
			}
			outputBuf.WriteString(line + "\n")
		}
		errorCHan <- scanner.Err()
	}()

	// 启动命令
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("error starting command: %v", err)
	} else if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("error waiting for command: %v", err)
	} else {

		for range 3 {
			if err := <-errorCHan; err != nil {
				return "", fmt.Errorf("error waiting for command: %v", err)
			}
		}

		return outputBuf.String(), nil
	}
}

func Command(command string) (string, error) {
	return runCommand(command, &commandOption{
		Stdout:       os.Stdout,
		Stderr:       os.Stderr,
		Sudo:         false,
		SudoPassword: "",
	})
}

func SudoCommand(command string, password string) (string, error) {
	return runCommand(command, &commandOption{
		Stdout:       os.Stdout,
		Stderr:       os.Stderr,
		Sudo:         true,
		SudoPassword: password,
	})
}
