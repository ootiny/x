package x

import (
	"bytes"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type sshOutput struct {
	contentChangeCH chan string
	buf             *bytes.Buffer
	mu              *sync.Mutex
}

func newSSHOutput(useExpect bool) *sshOutput {
	ret := &sshOutput{
		buf: bytes.NewBuffer(nil),
		mu:  &sync.Mutex{},
	}

	if useExpect {
		ret.contentChangeCH = make(chan string)
	}

	return ret
}

func (p *sshOutput) Write(data []byte) (n int, err error) {
	n, err = p.buf.Write(data)

	if ch := p.GetChangeCH(); ch != nil {
		p.contentChangeCH <- p.buf.String()
	}

	return n, err
}

func (p *sshOutput) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.buf.String()
}

func (p *sshOutput) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.contentChangeCH != nil {
		close(p.contentChangeCH)
		p.contentChangeCH = nil
	}
	return nil
}

func (p *sshOutput) GetChangeCH() chan string {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.contentChangeCH
}

func (p *sshOutput) WaitChange() (string, error) {
	if ch := p.GetChangeCH(); ch == nil {
		return "", io.EOF
	} else if v, ok := <-ch; ok {
		return v, nil
	} else {
		return "", io.EOF
	}
}

type SSHOption struct {
	Expect     func(output string) (string, error)
	Stdout     io.Writer
	Stderr     io.Writer
	User       string
	Host       string
	Password   string
	PrivateKey string
	Timeout    time.Duration
}

func NewSSHOptionWithPassword(user, host, password string) *SSHOption {
	return &SSHOption{
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
		User:     user,
		Host:     host,
		Password: password,
	}
}

func NewSSHOptionWithPrivateKey(user, host, privateKey string) *SSHOption {
	return &SSHOption{
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		User:       user,
		Host:       host,
		PrivateKey: privateKey,
	}
}

func SSH(command string, option *SSHOption) (string, error) {
	if option.User == "" {
		return "", Errorf("user is empty")
	}
	if option.Host == "" {
		return "", Errorf("host is empty")
	}

	if option.Timeout == 0 {
		option.Timeout = time.Second * 60
	}

	auth := []ssh.AuthMethod{}
	if option.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(option.PrivateKey))
		if err != nil {
			return "", Errorf("failed to parse private key: %v", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	} else if option.Password != "" {
		auth = append(auth, ssh.Password(option.Password))
	} else {
		return "", Errorf("password or private key is empty")
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(option.Host, "22"), &ssh.ClientConfig{
		User:            option.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 在生产环境中不推荐
		Timeout:         option.Timeout,
	})
	if err != nil {
		return "", Errorf("failed to dial: %s@%s:22 : %v", option.User, option.Host, err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		return "", Errorf("error creating stdin pipe: %v", err)
	}
	defer stdin.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", Errorf("error creating stdout pipe: %v", err)
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return "", Errorf("error creating stderr pipe: %v", err)
	}

	outCH := make(chan error, 2)
	inCH := make(chan error, 1)
	output := bytes.NewBuffer(nil)
	expectOutput := newSSHOutput(option.Expect != nil)

	// build stdout
	outWriters := []io.Writer{output, expectOutput}
	if option.Stdout != nil {
		outWriters = append(outWriters, option.Stdout)
	}
	useStdout := io.MultiWriter(outWriters...)

	// build stderr
	errWriters := []io.Writer{output, expectOutput}
	if option.Stderr != nil {
		errWriters = append(errWriters, option.Stderr)
	}
	useStdErr := io.MultiWriter(errWriters...)

	go func() {
		if option.Expect != nil {
			for {
				outputStr, err := expectOutput.WaitChange()
				if err != nil {
					inCH <- err
					return
				}

				if input, err := option.Expect(outputStr); err != nil {
					inCH <- err
					return
				} else if input != "" {
					if _, err := Fprint(stdin, input); err != nil {
						inCH <- err
						return
					}
				} else {
					continue
				}
			}
		} else {
			inCH <- io.EOF
		}
	}()

	go func() {
		_, err := io.Copy(useStdout, stdout)
		outCH <- err
	}()

	go func() {
		_, err := io.Copy(useStdErr, stderr)
		outCH <- err
	}()

	ColorPrintf("purple", "%s@%s: ", option.User, option.Host)
	ColorPrintf("blue", "%s\n", command)

	retError := session.Run(command)

	for range 2 {
		if err := <-outCH; err != nil && err != io.EOF {
			retError = err
		}
	}

	expectOutput.Close()

	if err := <-inCH; err != nil && err != io.EOF {
		retError = err
	}

	if retError != nil {
		return "", retError
	} else {
		return output.String(), nil
	}
}
