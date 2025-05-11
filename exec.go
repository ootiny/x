package x

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
)

type DoubleWriter struct {
	l io.Writer
	r io.Writer
}

func (w *DoubleWriter) Write(p []byte) (n int, err error) {
	n = 0
	err = nil
	if w.l != nil {
		n, err = w.l.Write(p)
		if err != nil {
			return
		}
	}

	if w.r != nil {
		n, err = w.r.Write(p)
		if err != nil {
			return
		}
	}

	return
}

type commandSplitItem struct {
	op    byte
	value string
}

func commandSplitString(s string, charArray []byte) []commandSplitItem {
	var ret []commandSplitItem
	var currentPart strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	lastOp := byte(0)

	for i := range len(s) {
		ch := s[i]

		// 处理引号
		if ch == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			currentPart.WriteByte(ch)
		} else if ch == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			currentPart.WriteByte(ch)
		} else if slices.Contains(charArray, ch) && !inSingleQuote && !inDoubleQuote {
			ret = append(ret, commandSplitItem{
				op:    lastOp,
				value: strings.TrimSpace(currentPart.String()),
			})
			lastOp = ch
			currentPart.Reset()
		} else {
			currentPart.WriteByte(ch)
		}
	}

	// 添加最后一部分
	if currentPart.Len() > 0 {
		ret = append(ret, commandSplitItem{
			op:    lastOp,
			value: strings.TrimSpace(currentPart.String()),
		})
	}

	return ret
}

type CommandConfig struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type Command struct {
	config *CommandConfig
}

func NewCommand(option ...*CommandConfig) *Command {
	if len(option) > 1 {
		panic("only one option is allowed")
	} else if len(option) == 1 {
		return &Command{
			config: option[0],
		}
	} else {
		return &Command{
			config: &CommandConfig{
				Stdin:  os.Stdin,
				Stdout: os.Stdout,
				Stderr: os.Stderr,
			},
		}
	}
}

func (c *Command) Eval(format string, args ...any) (string, error) {
	evalString := strings.TrimSpace(Sprintf(format, args...))

	evalCommnads := commandSplitString(evalString, []byte{'|'})

	if len(evalCommnads) == 0 {
		return "", fmt.Errorf("command cannot be empty")
	}

	lastResult := ""

	for idx, cmd := range evalCommnads {
		parts := strings.Fields(cmd.value)
		if len(parts) == 0 {
			return "", fmt.Errorf("command cannot be empty")
		}

		config := (*CommandConfig)(nil)

		if len(evalCommnads) == 1 {
			config = c.config
		} else if idx == 0 {
			config = &CommandConfig{
				Stdin:  c.config.Stdin,
				Stdout: nil,
				Stderr: c.config.Stderr,
			}
		} else if idx == len(evalCommnads)-1 {
			config = &CommandConfig{
				Stdin:  strings.NewReader(lastResult),
				Stdout: c.config.Stdout,
				Stderr: c.config.Stderr,
			}
		} else {
			config = &CommandConfig{
				Stdin:  strings.NewReader(lastResult),
				Stdout: nil,
				Stderr: c.config.Stderr,
			}
		}

		if result, err := runCommand(config, parts); err != nil {
			return "", err
		} else {
			lastResult = result
		}
	}

	return lastResult, nil
}

func runCommand(config *CommandConfig, parts []string) (string, error) {
	cmd := exec.Command(parts[0], parts[1:]...)

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

	output := bytes.NewBuffer(nil)
	errorCHan := make(chan error, 3)

	go func() {
		defer stdin.Close()
		_, err := io.Copy(stdin, config.Stdin)
		errorCHan <- err
	}()

	go func() {
		defer stdout.Close()
		_, err := io.Copy(&DoubleWriter{l: config.Stdout, r: output}, stdout)
		errorCHan <- err
	}()

	go func() {
		defer stderr.Close()
		if config.Stderr != nil {
			_, err := io.Copy(config.Stderr, stderr)
			errorCHan <- err
		} else {
			_, err := io.Copy(os.Stderr, stderr)
			errorCHan <- err
		}
	}()

	// 启动命令
	if err := cmd.Start(); err != nil {
		return "", Errorf("error starting command: %v", err)
	} else {
		for range 3 {
			if err := <-errorCHan; err != nil && err != io.EOF {
				return "", err
			}
		}
		return output.String(), nil
	}
}
