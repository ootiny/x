package x

import (
	"bytes"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	User         string
	Host         string
	Port         uint16
	Password     string
	PrivateKey   string
	SSHTimeoutMS uint32
	SCPTimeoutMS uint32
}

// progressWriter wraps an io.Writer and reports progress
type scpProgressWriter struct {
	writer     io.Writer
	total      int64
	current    int64
	lastUpdate time.Time
	onProgress func(current, total int64)
}

func (pw *scpProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	pw.current += int64(n)

	// Update progress at most 10 times per second to avoid flooding the terminal
	if time.Since(pw.lastUpdate) > 100*time.Millisecond {
		pw.onProgress(pw.current, pw.total)
		pw.lastUpdate = time.Now()
	}

	return n, err
}

type newlineWriter struct {
	writer    io.Writer
	firstData bool
}

func WrapNewLineWriter(writer io.Writer) io.Writer {
	return &newlineWriter{
		writer:    writer,
		firstData: true,
	}
}

func (p *newlineWriter) Write(data []byte) (n int, err error) {
	if p.firstData {
		p.firstData = false
		_, _ = p.writer.Write([]byte("\n"))
	}
	return p.writer.Write(data)
}

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

type SSHResult struct {
	stdout string
	stderr string
	err    error
}

func (p *SSHResult) IsSuccess() bool {
	return p.err == nil
}

func (p *SSHResult) IsFailure() bool {
	return p.err != nil
}

func (p *SSHResult) StdoutContains(text string) bool {
	return strings.Contains(p.stdout, text)
}

func (p *SSHResult) StderrContains(text string) bool {
	return strings.Contains(p.stderr, text)
}

func (p *SSHResult) Stdout() string {
	return p.stdout
}

func (p *SSHResult) Stderr() string {
	return p.stderr
}

func (p *SSHResult) Error() error {
	return p.err
}

// SSHClient is a client for SSH connections
type SSHClient struct {
	config     SSHConfig
	runClient  *ssh.Client
	expect     func(output string) (string, error)
	stdout     io.Writer
	stderr     io.Writer
	sshTimeout time.Duration
	scpTimeout time.Duration
	sshTempDir string
	auth       []ssh.AuthMethod
	runMu      *sync.Mutex
	errorsMu   *sync.Mutex
	errors     []error
}

// NewSSHClient creates a new SSHClient
func NewSSHClient(config SSHConfig) *SSHClient {
	ret := &SSHClient{
		expect:     nil,
		stdout:     os.Stdout,
		stderr:     os.Stderr,
		config:     config,
		sshTimeout: time.Second * 60,
		scpTimeout: time.Second * 600,
		sshTempDir: "/tmp",
		auth:       []ssh.AuthMethod{},
		runMu:      &sync.Mutex{},
		errorsMu:   &sync.Mutex{},
		errors:     []error{},
	}

	if ret.config.Port == 0 {
		ret.config.Port = 22
	}

	if ret.config.Password != "" {
		ret.auth = append(ret.auth, ssh.Password(ret.config.Password))
	}

	if ret.config.PrivateKey != "" {
		ret.AuthPrivateKey(ret.config.PrivateKey)
	}

	if ret.config.SSHTimeoutMS > 0 {
		ret.sshTimeout = time.Duration(ret.config.SSHTimeoutMS) * time.Millisecond
	}

	if ret.config.SCPTimeoutMS > 0 {
		ret.scpTimeout = time.Duration(ret.config.SCPTimeoutMS) * time.Millisecond
	}

	return ret
}

func (p *SSHClient) setError(err error) *SSHClient {
	p.errorsMu.Lock()
	defer p.errorsMu.Unlock()
	p.errors = append(p.errors, err)
	return p
}

func (p *SSHClient) getLastError() error {
	p.errorsMu.Lock()
	defer p.errorsMu.Unlock()
	if len(p.errors) > 0 {
		return p.errors[len(p.errors)-1]
	}
	return nil
}

// GetUser returns the user for the SSHClient
func (p *SSHClient) GetUser() string {
	return p.config.User
}

// GetHost returns the host for the SSHClient
func (p *SSHClient) GetHost() string {
	return p.config.Host
}

// GetPort returns the port for the SSHClient
func (p *SSHClient) GetPort() uint16 {
	return p.config.Port
}

// SetSSHTempDir sets the temporary directory for the SSHClient
func (p *SSHClient) SetSSHTempDir(dir string) *SSHClient {
	p.runMu.Lock()
	defer p.runMu.Unlock()
	p.sshTempDir = dir
	return p
}

func (p *SSHClient) SetStdout(stdout io.Writer) *SSHClient {
	p.runMu.Lock()
	defer p.runMu.Unlock()
	p.stdout = stdout
	return p
}

func (p *SSHClient) SetStderr(stderr io.Writer) *SSHClient {
	p.runMu.Lock()
	defer p.runMu.Unlock()
	p.stderr = stderr
	return p
}

// SetExpect sets the expect function for the SSHClient
func (p *SSHClient) SetExpect(expect func(output string) (string, error)) *SSHClient {
	p.runMu.Lock()
	defer p.runMu.Unlock()

	if p.runClient != nil {
		panic("client is already open")
	}

	p.expect = expect
	return p
}

// SetSSHTimeout sets the timeout for the SSH connection
func (p *SSHClient) SetSSHTimeout(timeout time.Duration) *SSHClient {
	p.runMu.Lock()
	defer p.runMu.Unlock()

	if p.runClient != nil {
		panic("client is already open")
	}

	p.sshTimeout = timeout
	return p
}

// SetSCPTimeout sets the timeout for the SCP connection
func (p *SSHClient) SetSCPTimeout(timeout time.Duration) *SSHClient {
	p.runMu.Lock()
	defer p.runMu.Unlock()

	if p.runClient != nil {
		panic("client is already open")
	}

	p.scpTimeout = timeout
	return p
}

// AuthKeyboardInteractive adds a keyboard interactive authentication method to the SSHClient
func (p *SSHClient) AuthKeyboardInteractive(callback ssh.KeyboardInteractiveChallenge) *SSHClient {
	p.runMu.Lock()
	defer p.runMu.Unlock()

	if p.runClient != nil {
		panic("client is already open")
	}

	p.auth = append(p.auth, ssh.KeyboardInteractive(callback))
	return p
}

// AuthPrivateKey adds a private key authentication method to the SSHClient
func (p *SSHClient) AuthPrivateKey(privateKey string) *SSHClient {
	p.runMu.Lock()
	defer p.runMu.Unlock()

	if p.runClient != nil {
		panic("client is already open")
	}

	privateKeyPath := privateKey
	if strings.HasPrefix(privateKeyPath, "~/") {
		privateKeyPath = filepath.Join(os.Getenv("HOME"), privateKeyPath[2:])
	}

	// if privateKey as file exists, read the file
	if fileInfo, err := os.Stat(privateKeyPath); err == nil && !fileInfo.IsDir() {
		if v, err := os.ReadFile(privateKeyPath); err != nil {
			p.setError(err)
			return p
		} else {
			privateKey = string(v)
		}
	}

	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		p.setError(err)
		return p
	}
	p.auth = append(p.auth, ssh.PublicKeys(signer))
	return p
}

func (p *SSHClient) OpenWithRetry(retry int) error {
	p.runMu.Lock()
	defer p.runMu.Unlock()

	if retry <= 0 {
		retry = 1
	}

	if err := p.getLastError(); err != nil {
		return err
	}

	if p.config.User == "" {
		reportErr := Errorf("user is empty")
		p.setError(reportErr)
		return reportErr
	}
	if p.config.Host == "" {
		reportErr := Errorf("host is empty")
		p.setError(reportErr)
		return reportErr
	}

	var retError error

	for range retry {
		if client, err := ssh.Dial(
			"tcp",
			net.JoinHostPort(p.config.Host, Sprintf("%d", p.config.Port)),
			&ssh.ClientConfig{
				User:            p.config.User,
				Auth:            p.auth,
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				Timeout:         p.sshTimeout,
			},
		); err != nil {
			retError = Errorf("failed to dial: %s@%s:%d : %v", p.config.User, p.config.Host, p.config.Port, err)
			time.Sleep(time.Second * 1)
			continue
		} else {
			p.runClient = client
			return nil
		}
	}

	p.setError(retError)
	return retError
}

func (p *SSHClient) Open() error {
	p.runMu.Lock()
	defer p.runMu.Unlock()

	if err := p.getLastError(); err != nil {
		return err
	}

	if p.config.User == "" {
		reportErr := Errorf("user is empty")
		p.setError(reportErr)
		return reportErr
	}
	if p.config.Host == "" {
		reportErr := Errorf("host is empty")
		p.setError(reportErr)
		return reportErr
	}

	if client, err := ssh.Dial(
		"tcp",
		net.JoinHostPort(p.config.Host, Sprintf("%d", p.config.Port)),
		&ssh.ClientConfig{
			User:            p.config.User,
			Auth:            p.auth,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         p.sshTimeout,
		},
	); err != nil {
		reportErr := Errorf("failed to dial: %s@%s:%d : %v", p.config.User, p.config.Host, p.config.Port, err)
		p.setError(reportErr)
		return reportErr
	} else {
		p.runClient = client
		return nil
	}
}

func (p *SSHClient) Close() error {
	p.runMu.Lock()
	defer p.runMu.Unlock()

	if p.runClient != nil {
		ret := p.runClient.Close()
		if ret != nil {
			return ret
		} else {
			p.runClient = nil
			return nil
		}
	} else {
		return nil
	}
}

func (p *SSHClient) RemoteHomeDir() (string, error) {
	if result := p.SSH("echo $HOME"); result.IsSuccess() {
		return result.Stdout(), nil
	} else {
		return "", result.Error()
	}
}

func (p *SSHClient) SudoSSH(format string, args ...any) *SSHResult {
	if p.config.User == "root" {
		return p.ssh(false, format, args...)
	} else {
		return p.ssh(true, format, args...)
	}
}

func (p *SSHClient) SSH(format string, args ...any) *SSHResult {
	return p.ssh(false, format, args...)
}

// SSH executes a command on the SSHClient
func (p *SSHClient) ssh(sudo bool, format string, args ...any) *SSHResult {
	p.runMu.Lock()
	defer p.runMu.Unlock()

	if err := p.getLastError(); err != nil {
		return &SSHResult{
			err: err,
		}
	}

	if p.runClient == nil {
		reportErr := Errorf("client is not open")
		p.setError(reportErr)
		return &SSHResult{
			err: reportErr,
		}
	}

	command := Sprintf(format, args...)
	if sudo {
		command = Sprintf("sudo -S %s", command)
	}

	session, err := p.runClient.NewSession()
	if err != nil {
		reportErr := Errorf("failed to create session: %v", err)
		p.setError(reportErr)
		return &SSHResult{
			err: reportErr,
		}
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		reportErr := Errorf("error creating stdin pipe: %v", err)
		p.setError(reportErr)
		return &SSHResult{
			err: reportErr,
		}
	}
	defer stdin.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		reportErr := Errorf("error creating stdout pipe: %v", err)
		p.setError(reportErr)
		return &SSHResult{
			err: reportErr,
		}
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		reportErr := Errorf("error creating stderr pipe: %v", err)
		p.setError(reportErr)
		return &SSHResult{
			err: reportErr,
		}
	}

	outCH := make(chan error, 2)
	inCH := make(chan error, 1)
	outBuffer := bytes.NewBuffer(nil)
	errBuffer := bytes.NewBuffer(nil)
	expectOutput := newSSHOutput(p.expect != nil)

	// build stdout
	outWriters := []io.Writer{outBuffer, expectOutput}
	if p.stdout != nil {
		outWriters = append(outWriters, WrapNewLineWriter(p.stdout))
	}
	useStdout := io.MultiWriter(outWriters...)

	// build stderr
	errWriters := []io.Writer{errBuffer, expectOutput}
	if p.stderr != nil {
		errWriters = append(errWriters, WrapNewLineWriter(p.stderr))
	}
	useStdErr := io.MultiWriter(errWriters...)

	go func() {
		if p.expect != nil {
			for {
				outputStr, err := expectOutput.WaitChange()
				if err != nil {
					inCH <- err
					return
				}

				if input, err := p.expect(outputStr); err != nil {
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

	ColorPrintf("purple", "%s@%s: ", p.config.User, p.config.Host)
	ColorPrintf("blue", "%s", command)

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
		if p.stderr == nil {
			ColorPrintf("red", "\n✗ failed\n")
		} else if !strings.HasSuffix(expectOutput.String(), "\n") {
			Print("\n")
		} else {
			Ignore()
		}

		return &SSHResult{
			stdout: strings.TrimSpace(outBuffer.String()),
			stderr: strings.TrimSpace(errBuffer.String()),
			err:    retError,
		}
	} else {
		if p.stdout == nil {
			ColorPrintf("green", "\n✔ ok\n")
		} else if !strings.HasSuffix(outBuffer.String(), "\n") {
			Print("\n")
		} else {
			Ignore()
		}

		return &SSHResult{
			stdout: strings.TrimSpace(outBuffer.String()),
			stderr: strings.TrimSpace(errBuffer.String()),
			err:    nil,
		}
	}
}

// SCP uploads a file to the SSHClient
func (p *SSHClient) scp(localPath string, remotePath string) error {
	p.runMu.Lock()
	defer p.runMu.Unlock()

	if p.config.User == "" {
		return Errorf("user is empty")
	}
	if p.config.Host == "" {
		return Errorf("host is empty")
	}

	if strings.HasPrefix(localPath, "~/") {
		localPath = filepath.Join(os.Getenv("HOME"), localPath[2:])
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

	client, err := ssh.Dial(
		"tcp",
		net.JoinHostPort(p.config.Host, Sprintf("%d", p.config.Port)),
		&ssh.ClientConfig{
			User:            p.config.User,
			Auth:            p.auth,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         p.scpTimeout,
		},
	)
	if err != nil {
		return Errorf("failed to dial: %s@%s:%d : %v", p.config.User, p.config.Host, p.config.Port, err)
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
		progressWriter := &scpProgressWriter{
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
	ColorPrintf("purple", "%s@%s:", p.config.User, p.config.Host)
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

func (p *SSHClient) IsFileExists(filePath string) (bool, error) {
	if result := p.SudoSSH("test -f %s && echo 'yes' || echo 'no'", filePath); result.IsSuccess() {
		return result.Stdout() == "yes", nil
	} else {
		return false, result.Error()
	}
}

func (p *SSHClient) IsDirectoryExists(dirPath string) (bool, error) {
	if result := p.SudoSSH("test -d %s && echo 'yes' || echo 'no'", dirPath); result.IsSuccess() {
		return result.Stdout() == "yes", nil
	} else {
		return false, result.Error()
	}
}

func (p *SSHClient) CreateDirectory(dirPath string, user string, group string, mode os.FileMode) error {
	if exists, err := p.IsDirectoryExists(dirPath); err != nil {
		return err
	} else if exists {
		return nil
	} else {
		// create parent directory
		parentDir := filepath.Dir(dirPath)
		if existsParent, err := p.IsDirectoryExists(parentDir); err != nil {
			return err
		} else if !existsParent {
			if err := p.CreateDirectory(parentDir, user, group, mode); err != nil {
				return err
			} else {
				Ignore()
			}
		} else {
			Ignore()
		}

		// create directory
		if result := p.SudoSSH("mkdir -p %s", dirPath); result.IsFailure() {
			return result.Error()
		} else if result := p.SudoSSH("chown %s:%s %s", user, group, dirPath); result.IsFailure() {
			return result.Error()
		} else if result := p.SudoSSH("chmod %o %s", mode, dirPath); result.IsFailure() {
			return result.Error()
		} else {
			return nil
		}
	}
}

func (p *SSHClient) SCPFile(
	localPath string, remotePath string,
	user string, group string, mode os.FileMode,
) error {
	tmpName := RandFileName(16) + ".tmp"
	remoteTempPath := filepath.Join(p.sshTempDir, tmpName)

	if err := p.CreateDirectory(filepath.Dir(remotePath), user, group, 0755); err != nil {
		return err
	} else if err := p.scp(localPath, remoteTempPath); err != nil {
		return err
	} else if result := p.SudoSSH("mv %s %s", remoteTempPath, remotePath); result.IsFailure() {
		return result.Error()
	} else if result := p.SudoSSH("chown %s:%s %s", user, group, remotePath); result.IsFailure() {
		return result.Error()
	} else if result := p.SudoSSH("chmod %o %s", mode, remotePath); result.IsFailure() {
		return result.Error()
	} else {
		return nil
	}
}

func (p *SSHClient) SCPBytes(
	bytes []byte, remotePath string,
	user string, group string, mode os.FileMode,
) error {
	tmpName := RandFileName(16) + ".tmp"

	// write bytes to  file tmpName
	tempFile, err := os.Create(tmpName)
	if err != nil {
		return Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()
	defer os.Remove(tmpName)

	if _, err := tempFile.Write(bytes); err != nil {
		return Errorf("failed to write bytes to temp file: %w", err)
	}

	return p.SCPFile(tmpName, remotePath, user, group, mode)
}

// IsLinuxServiceEnabled checks if a service is enabled
func (p *SSHClient) IsLinuxServiceEnabled(serviceName string) (bool, error) {
	result := p.SudoSSH("systemctl is-enabled %s", serviceName)

	if result.StdoutContains("enabled") {
		return true, nil
	} else if result.StdoutContains("disabled") {
		return false, nil
	} else {
		return false, result.Error()
	}
}

// IsLinuxServiceRunning checks if a service is running
func (p *SSHClient) IsLinuxServiceRunning(serviceName string) (bool, error) {
	result := p.SudoSSH("systemctl is-active %s", serviceName)

	if result.StdoutContains("inactive") || result.StdoutContains("failed") {
		return false, nil
	} else if result.StdoutContains("active") {
		return true, nil
	} else {
		return false, result.Error()
	}
}

// StopLinuxService stops a service
func (p *SSHClient) StopLinuxService(serviceName string) error {
	if running, err := p.IsLinuxServiceRunning(serviceName); err != nil {
		return err
	} else if !running {
		return nil
	} else if result := p.SudoSSH("systemctl stop %s", serviceName); result.IsFailure() {
		return result.Error()
	} else {
		return nil
	}
}

// DisableLinuxService disables a service
func (p *SSHClient) DisableLinuxService(serviceName string) error {
	if enabled, err := p.IsLinuxServiceEnabled(serviceName); err != nil {
		return err
	} else if !enabled {
		return nil
	} else if result := p.SudoSSH("systemctl disable %s", serviceName); result.IsFailure() {
		return result.Error()
	} else {
		return nil
	}
}

// EnableLinuxService enables a service
func (p *SSHClient) EnableLinuxService(serviceName string) error {
	if enabled, err := p.IsLinuxServiceEnabled(serviceName); err != nil {
		return err
	} else if enabled {
		return nil
	} else if result := p.SudoSSH("systemctl enable %s", serviceName); result.IsFailure() {
		return result.Error()
	} else {
		return nil
	}
}

// StartLinuxService starts a service
func (p *SSHClient) StartLinuxService(serviceName string) error {
	if running, err := p.IsLinuxServiceRunning(serviceName); err != nil {
		return err
	} else if running {
		return nil
	} else if result := p.SudoSSH("systemctl start %s", serviceName); result.IsFailure() {
		return result.Error()
	} else {
		return nil
	}
}

// DeployLinuxService deploys a linux service
func (p *SSHClient) DeployLinuxService(
	serviceContent string,
	serviceRemoteFilePath string,
) error {
	serviceName := filepath.Base(serviceRemoteFilePath)

	if len(strings.TrimSpace(serviceContent)) > 0 {
		if err := p.SCPBytes([]byte(serviceContent), serviceRemoteFilePath, "root", "root", 0644); err != nil {
			return err
		}
	}

	if result := p.SudoSSH("systemctl daemon-reload"); result.IsFailure() {
		return result.Error()
	} else if err := p.DisableLinuxService(serviceName); err != nil {
		return result.Error()
	} else if err := p.EnableLinuxService(serviceName); err != nil {
		return result.Error()
	} else if err := p.StopLinuxService(serviceName); err != nil {
		return err
	} else if err := p.StartLinuxService(serviceName); err != nil {
		return err
	} else {
		return nil
	}
}

func (p *SSHClient) GetLinuxArch() (string, error) {
	if result := p.SSH("uname -m"); result.IsFailure() {
		return "", result.Error()
	} else if result.StdoutContains("aarch64") || result.StdoutContains("arm64") {
		return "arm64", nil
	} else if result.StdoutContains("x86_64") || result.StdoutContains("amd64") {
		return "amd64", nil
	} else {
		return "", Errorf("unsupported platform: %s", result.Stdout())
	}
}
