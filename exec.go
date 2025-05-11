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

type commandWriteFileList struct {
	paths []string
	files []*os.File
}

func (p *commandWriteFileList) IOWriterList() []io.Writer {
	writers := make([]io.Writer, len(p.files))
	for idx, file := range p.files {
		writers[idx] = file
	}
	return writers
}

func (p *commandWriteFileList) IOReaderList() []io.Reader {
	readers := make([]io.Reader, len(p.files))
	for idx, file := range p.files {
		readers[idx] = file
	}
	return readers
}

func newCommandWriteFileList(paths []string) *commandWriteFileList {
	return &commandWriteFileList{
		paths: paths,
		files: nil,
	}
}

func (p *commandWriteFileList) Open(readOnly bool) error {
	if p.files == nil {
		p.files = make([]*os.File, len(p.paths))

		for idx, path := range p.paths {
			if readOnly {
				if f, err := os.Open(path); err != nil {
					return err
				} else {
					p.files[idx] = f
				}
			} else {
				if f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err != nil {
					return err
				} else {
					p.files[idx] = f
				}
			}
		}

		return nil
	} else {
		return fmt.Errorf("files already opened")
	}
}

func (p *commandWriteFileList) Close() error {
	err := error(nil)

	if p.files != nil {
		for _, file := range p.files {
			if e := file.Close(); e != nil {
				err = e
			}
		}
	}

	return err
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

func (c *Command) SetStdin(stdin io.Reader) {
	c.config.Stdin = stdin
}

func (c *Command) SetStdout(stdout io.Writer) {
	c.config.Stdout = stdout
}

func (c *Command) SetStderr(stderr io.Writer) {
	c.config.Stderr = stderr
}

func (c *Command) Eval(format string, args ...any) (string, error) {
	evalString := strings.TrimSpace(Sprintf(format, args...))

	evalCommnads := commandSplitString(evalString, []byte{'|'})

	if len(evalCommnads) == 0 {
		return "", fmt.Errorf("command cannot be empty")
	}

	lastResult := ""

	for idx, cmd := range evalCommnads {
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

		if result, err := runCommand(config, cmd.value); err != nil {
			return "", err
		} else {
			lastResult = result
		}
	}

	return lastResult, nil
}

func runCommand(config *CommandConfig, command string) (string, error) {

	commandList := commandSplitString(command, []byte{'<', '>'})

	if len(commandList) == 0 {
		return "", fmt.Errorf("command cannot be empty")
	}

	parts := strings.Fields(commandList[0].value)

	if len(parts) == 0 {
		return "", fmt.Errorf("command cannot be empty")
	}

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

	inputFiles := []string{}
	outputFiles := []string{}

	for _, cmd := range commandList {
		if cmd.op == '<' {
			inputFiles = append(inputFiles, cmd.value)
		} else if cmd.op == '>' {
			outputFiles = append(outputFiles, cmd.value)
		} else {
			continue
		}
	}

	// build stdin
	useStdin := config.Stdin
	if len(inputFiles) > 0 {
		list := newCommandWriteFileList(inputFiles)
		if err := list.Open(true); err != nil {
			return "", Errorf("error opening input files: %v", err)
		}
		defer list.Close()
		useStdin = io.MultiReader(list.IOReaderList()...)
	}

	// build stdout
	writers := make([]io.Writer, 0)
	writers = append(writers, output)
	if config.Stdout != nil {
		writers = append(writers, config.Stdout)
	}
	if len(outputFiles) > 0 {
		list := newCommandWriteFileList(outputFiles)
		if err := list.Open(false); err != nil {
			return "", Errorf("error opening output files: %v", err)
		}
		defer list.Close()
		writers = append(writers, list.IOWriterList()...)

	}
	useStdout := io.MultiWriter(writers...)

	// build stderr
	useStderr := config.Stderr

	go func() {
		defer stdin.Close()
		if useStdin != nil {
			_, err := io.Copy(stdin, useStdin)
			errorCHan <- err
		} else {
			errorCHan <- nil
		}
	}()

	go func() {
		defer stdout.Close()
		_, err := io.Copy(useStdout, stdout)
		errorCHan <- err
	}()

	go func() {
		defer stderr.Close()
		if useStderr != nil {
			_, err := io.Copy(useStderr, stderr)
			errorCHan <- err
		} else {
			errorCHan <- nil
		}
	}()

	// 启动命令
	if err := cmd.Start(); err != nil {
		return "", Errorf("error starting command: %v", err)
	} else {
		// wait for stdout and stderr
		for range 3 {
			if err := <-errorCHan; err != nil && err != io.EOF {
				return "", err
			}
		}

		return output.String(), nil
	}
}
