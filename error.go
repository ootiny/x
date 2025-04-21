package x

import "fmt"

// Errorf formats according to a format specifier and returns the string as a
// value that satisfies error.
func Errorf(format string, a ...any) error {
	return fmt.Errorf(format, a...)
}
