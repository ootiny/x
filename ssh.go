package x

import (
	"bytes"
	"io"
	"net"
	"os"
	"path/filepath"
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
	Port       string
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
		Port:     "22",
		Password: password,
	}
}

func NewSSHOptionWithPrivateKey(user, host, privateKey string) *SSHOption {
	return &SSHOption{
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		User:       user,
		Host:       host,
		Port:       "22",
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

	client, err := ssh.Dial("tcp", net.JoinHostPort(option.Host, option.Port), &ssh.ClientConfig{
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

func SCP(localPath string, remotePath string, option *SSHOption) error {
	if option.User == "" {
		return Errorf("user is empty")
	}
	if option.Host == "" {
		return Errorf("host is empty")
	}

	if option.Timeout == 0 {
		option.Timeout = time.Second * 600
	}

	file, err := os.Open(localPath)
	if err != nil {
		return Errorf("failed to open local file %s: %v", localPath, err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return Errorf("failed to stat local file %s: %v", localPath, err)
	}
	fileSize := stat.Size()
	fileName := filepath.Base(remotePath) // Use the base of remotePath as the filename

	auth := []ssh.AuthMethod{}
	if option.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(option.PrivateKey))
		if err != nil {
			return Errorf("failed to parse private key: %v", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	} else if option.Password != "" {
		auth = append(auth, ssh.Password(option.Password))
	} else {
		return Errorf("password or private key is empty")
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(option.Host, option.Port), &ssh.ClientConfig{
		User:            option.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 在生产环境中不推荐
		Timeout:         option.Timeout,
	})
	if err != nil {
		return Errorf("failed to dial: %s@%s:22 : %v", option.User, option.Host, err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		return Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return Errorf("failed to get stdout pipe: %w", err)
	}

	go func() {
		defer stdin.Close() // Crucial to close stdin to signal EOF to remote scp

		// SCP protocol:
		// 1. Send 'C' mode size filename\n
		//    Example: C0644 123 example.txt\n
		//    Mode 0644 is common for files.
		_, _ = Fprintf(stdin, "C0644 %d %s\n", fileSize, fileName)

		// Check for acknowledgment (a null byte) from remote scp
		ack := make([]byte, 1)
		_, err := stdout.Read(ack)
		if err != nil {
			LogErrorf("Error reading initial ack from remote scp: %v (stdout may contain error message)", err)
			// Attempt to read more for an error message from scp
			errMsg := make([]byte, 512)
			n, _ := stdout.Read(errMsg)
			if n > 0 {
				LogErrorf("Remote scp stdout/stderr: %s", string(errMsg[:n]))
			}
			return // Don't proceed if ack fails
		}
		if ack[0] != 0 {
			LogErrorf("Remote scp sent non-zero ack: %d. Error on remote.", ack[0])
			// Attempt to read more for an error message from scp
			errMsg := make([]byte, 512)
			n, _ := stdout.Read(errMsg)
			if n > 0 {
				LogErrorf("Remote scp stdout/stderr: %s", string(errMsg[:n]))
			}
			return
		}

		// 2. Send file contents with progress tracking
		progressWriter := &progressWriter{
			writer: stdin,
			total:  fileSize,
			onProgress: func(current, total int64) {
				percentage := float64(current) / float64(total) * 100
				ColorPrintf("cyan", "\rUploading: %.2f%% (%d/%d bytes)", percentage, current, total)
			},
		}

		copied, err := io.Copy(progressWriter, file)
		// Always show 100% at the end
		ColorPrintf("green", "\rUploading: 100.00%% (%d/%d bytes)", fileSize, fileSize)
		ColorPrintf("green", " - Finished!\n")

		if err != nil {
			LogErrorf("Error copying file contents to remote stdin: %v", err)
			return
		}
		if copied != fileSize {
			LogErrorf("Copied %d bytes, but expected %d bytes", copied, fileSize)
			return
		}

		// 3. Send a null byte to indicate EOF for this file
		_, _ = Fprint(stdin, "\x00")

		// 4. Check for final acknowledgment
		_, err = stdout.Read(ack)
		if err != nil {
			LogErrorf("Error reading final ack from remote scp: %v", err)
			return
		}
		if ack[0] != 0 {
			LogErrorf("Remote scp sent non-zero final ack: %d. Error on remote.", ack[0])
			errMsg := make([]byte, 512)
			n, _ := stdout.Read(errMsg)
			if n > 0 {
				LogErrorf("Remote scp stdout/stderr: %s", string(errMsg[:n]))
			}
			return
		}
	}()

	remoteTargetDir := filepath.Dir(remotePath)
	cmd := Sprintf("scp -t %s", remoteTargetDir)
	ColorPrintf("blue", "scp %s ", localPath)
	ColorPrintf("purple", "%s@%s:", option.User, option.Host)
	ColorPrintf("blue", "%s\n", remotePath)

	// session.Run() blocks until command finishes.
	// We need to use Start() because we are interacting with stdin/stdout.
	err = session.Start(cmd)
	if err != nil {
		return Errorf("failed to start remote scp command '%s': %w", cmd, err)
	}

	// Wait for the command to finish.
	// This will also wait for the goroutine above to complete its work with stdin/stdout.
	err = session.Wait()
	if err != nil {
		// Check if it's an ExitError to get more details
		if exitErr, ok := err.(*ssh.ExitError); ok {
			return Errorf("remote scp command failed with exit status %d: %w", exitErr.ExitStatus(), err)
		}
		return Errorf("remote scp command failed: %w", err)
	}

	return nil
}

// progressWriter wraps an io.Writer and reports progress
type progressWriter struct {
	writer     io.Writer
	total      int64
	current    int64
	lastUpdate time.Time
	onProgress func(current, total int64)
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	pw.current += int64(n)

	// Update progress at most 10 times per second to avoid flooding the terminal
	if time.Since(pw.lastUpdate) > 100*time.Millisecond {
		pw.onProgress(pw.current, pw.total)
		pw.lastUpdate = time.Now()
	}

	return n, err
}
