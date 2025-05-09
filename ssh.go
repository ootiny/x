package x

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHOption struct {
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
	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", Errorf("error creating stdout pipe: %v", err)
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return "", Errorf("error creating stderr pipe: %v", err)
	}

	defer stdin.Close()

	results := make(chan error, 2)
	var outputBuf bytes.Buffer

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if option.Stdout != nil {
				if _, err := Fprintln(option.Stdout, line); err != nil {
					results <- err
					return
				}
			}
			outputBuf.WriteString(line + "\n")
		}
		results <- scanner.Err()
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if option.Stderr != nil {
				if _, err := Fprintln(option.Stderr, line); err != nil {
					results <- err
					return
				}
			}
			outputBuf.WriteString(line + "\n")
		}
		results <- scanner.Err()
	}()

	retError := session.Run(command)

	for range 2 {
		err = <-results
		if retError == nil && err != nil {
			retError = err
		}
	}

	return outputBuf.String(), retError
}
