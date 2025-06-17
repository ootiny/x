package x

import (
	"strings"
)

type ExecResult struct {
	stdout string
	stderr string
	err    error
}

func (p *ExecResult) IsSuccess() bool {
	return p.err == nil
}

func (p *ExecResult) IsFailure() bool {
	return p.err != nil
}

func (p *ExecResult) StdoutContains(text string) bool {
	return strings.Contains(p.stdout, text)
}

func (p *ExecResult) StderrContains(text string) bool {
	return strings.Contains(p.stderr, text)
}

func (p *ExecResult) Stdout() string {
	return p.stdout
}

func (p *ExecResult) Stderr() string {
	return p.stderr
}

func (p *ExecResult) Error() error {
	return p.err
}

func Ternary[T any](cond bool, trueValue, falseValue T) T {
	if cond {
		return trueValue
	} else {
		return falseValue
	}
}
