package x

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"
)

type commandFileList struct {
	paths    []string
	files    []*os.File
	readOnly bool
	mu       *sync.Mutex
}

func newCommandFileList(paths []string, readOnly bool) *commandFileList {
	return &commandFileList{
		paths:    paths,
		files:    nil,
		readOnly: readOnly,
		mu:       &sync.Mutex{},
	}
}

func (p *commandFileList) IOWriterList() []io.Writer {
	p.mu.Lock()
	defer p.mu.Unlock()

	writers := make([]io.Writer, len(p.files))
	for idx, file := range p.files {
		writers[idx] = file
	}
	return writers
}

func (p *commandFileList) IOReaderList() []io.Reader {
	p.mu.Lock()
	defer p.mu.Unlock()

	readers := make([]io.Reader, len(p.files))
	for idx, file := range p.files {
		readers[idx] = file
	}
	return readers
}

func (p *commandFileList) Open() (err error) {
	defer func() {
		if err != nil {
			_ = p.Close()
		}
	}()

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.files == nil {
		p.files = make([]*os.File, len(p.paths))
	}

	for idx, path := range p.paths {
		if p.files[idx] != nil {
			return fmt.Errorf("file already opened")
		}

		if p.readOnly {
			if f, err := os.Open(path); err != nil {
				return err
			} else {
				p.files[idx] = f
			}
		} else {
			if f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644); err != nil {
				return err
			} else {
				p.files[idx] = f
			}
		}
	}

	return nil
}

func (p *commandFileList) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := error(nil)
	if p.files != nil {
		for idx, file := range p.files {
			if e := file.Close(); e != nil {
				err = e
			} else {
				p.files[idx] = nil
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
		if option[0].Stdin == os.Stdin {
			panic("stdin cannot be os.Stdin")
		}

		return &Command{
			config: option[0],
		}
	} else {
		return &Command{
			config: &CommandConfig{
				Stdin:  nil,
				Stdout: os.Stdout,
				Stderr: os.Stderr,
			},
		}
	}
}

func (c *Command) SetStdin(stdin io.Reader) {
	if stdin == os.Stdin {
		panic("stdin cannot be os.Stdin")
	}
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

	// build input and output files
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
	readers := make([]io.Reader, 0)
	if config.Stdin != nil {
		readers = append(readers, config.Stdin)
	}
	inputList := newCommandFileList(inputFiles, true)
	if err := inputList.Open(); err != nil {
		return "", Errorf("error opening input files: %v", err)
	}
	defer inputList.Close()
	readers = append(readers, inputList.IOReaderList()...)
	useStdin := io.MultiReader(readers...)

	// build stdout
	writers := make([]io.Writer, 0)
	writers = append(writers, output)
	if config.Stdout != nil {
		writers = append(writers, config.Stdout)
	}
	outputList := newCommandFileList(outputFiles, false)
	if err := outputList.Open(); err != nil {
		return "", Errorf("error opening output files: %v", err)
	}
	defer outputList.Close()
	writers = append(writers, outputList.IOWriterList()...)
	useStdout := io.MultiWriter(writers...)

	// build stderr
	useStderr := config.Stderr

	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		return "", Errorf("error starting command: %v", err)
	} else {
		go func() {
			err := error(nil)
			if useStdin != nil {
				_, err = io.Copy(stdin, useStdin)
			}
			stdin.Close()
			errorCHan <- err
		}()

		go func() {
			err := error(nil)
			if useStdout != nil {
				_, err = io.Copy(useStdout, stdout)
			}
			stdout.Close()
			inputList.Close()
			outputList.Close()
			errorCHan <- err
		}()

		go func() {
			err := error(nil)
			if useStderr != nil {
				_, err = io.Copy(useStderr, stderr)
			}
			stderr.Close()
			errorCHan <- err
		}()

		// wait for stdout and stderr
		for range 3 {
			if err := <-errorCHan; err != nil && err != io.EOF {
				return "", err
			}
		}

		return output.String(), nil
	}
}
