package x

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type commandOption struct {
	Stdout       io.Writer
	Stderr       io.Writer
	Sudo         bool
	SudoPassword string
}

func runCommand(command string, option *commandOption) (string, error) {
	if command == "" {
		return "", fmt.Errorf("command cannot be empty")
	}
	if option == nil {
		return "", fmt.Errorf("option is nil")
	}

	var cmd *exec.Cmd
	shellCommand := []string{"sh", "-c", command}

	if option.Sudo {
		cmd = exec.Command("sudo", append([]string{"-S"}, shellCommand...)...)
		// Provide the password via cmd.Stdin.
		// If SudoPassword is empty, sudo -S will receive an immediate EOF,
		// which usually causes it to fail or prompt on TTY if available.
		cmd.Stdin = strings.NewReader(option.SudoPassword + "\n")
	} else {
		cmd = exec.Command(shellCommand[0], shellCommand[1:]...)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("error creating stdout pipe: %v", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		// stdoutPipe was successfully created, but we are erroring out.
		// It's good practice to close resources we've opened.
		// However, os.File.Close on a pipe's read end doesn't prevent the child from writing;
		// it just means our side won't read. cmd.Wait() usually handles cleanup.
		// For simplicity here, we'll let it be. In more complex scenarios, explicit closing might be needed.
		return "", fmt.Errorf("error creating stderr pipe: %v", err)
	}

	var outputBuf bytes.Buffer
	var wg sync.WaitGroup
	var multiError error // To collect errors from goroutines and cmd.Wait()
	var errMu sync.Mutex // To protect multiError

	// Function to safely append errors
	addError := func(newErr error) {
		if newErr == nil {
			return
		}
		errMu.Lock()
		defer errMu.Unlock()
		if multiError == nil {
			multiError = newErr
		} else {
			multiError = fmt.Errorf("%v; %w", multiError, newErr)
		}
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("error starting command: %v", err)
	}

	// Goroutine to handle stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		var writers []io.Writer
		writers = append(writers, &outputBuf) // Always capture to outputBuf
		if option.Stdout != nil {
			writers = append(writers, option.Stdout) // Also write to provided Stdout
		}
		combinedStdout := io.MultiWriter(writers...)
		if _, copyErr := io.Copy(combinedStdout, stdoutPipe); copyErr != nil && copyErr != io.EOF {
			// This error means copying from the pipe failed, not that the command itself failed.
			// For example, if option.Stdout is a writer that errors.
			// Ignore EOF errors as they are expected when the pipe is closed
			addError(fmt.Errorf("error copying stdout: %w", copyErr))
		}
	}()

	// Goroutine to handle stderr
	wg.Add(1)
	go func() {
		defer wg.Done()
		var writers []io.Writer
		writers = append(writers, &outputBuf) // Always capture to outputBuf
		if option.Stderr != nil {
			writers = append(writers, option.Stderr) // Also write to provided Stderr
		}
		combinedStderr := io.MultiWriter(writers...)
		if _, copyErr := io.Copy(combinedStderr, stderrPipe); copyErr != nil && copyErr != io.EOF {
			// Ignore EOF errors as they are expected when the pipe is closed
			addError(fmt.Errorf("error copying stderr: %w", copyErr))
		}
	}()

	// Wait for the command to finish. This also ensures that the child process
	// has closed its ends of the stdout/stderr pipes.
	waitErr := cmd.Wait()
	addError(waitErr) // Add command execution error (like non-zero exit status)

	// Wait for all I/O goroutines to finish their copying.
	// This must happen after cmd.Wait() has returned, ensuring pipes are EOF.
	wg.Wait()

	return outputBuf.String(), multiError
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
