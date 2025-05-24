package x

import (
	"fmt"
	"io"
)

// WarpTabs wraps the string with tabs.
func WarpTabs(numbOfTab uint, s string) string {
	prifix := ""
	for range int(numbOfTab) {
		prifix += "\t"
	}
	return prifix + s
}

// Errorf formats according to a format specifier and returns the string as an error.
func Errorf(format string, a ...any) error {
	return fmt.Errorf(format, a...)
}

// Panicf formats according to a format specifier and panics with the formatted error.
func Panicf(format string, a ...any) {
	panic(Errorf(format, a...))
}

// Print formats using the default formats for its operands and writes to standard output.
func Print(a ...any) {
	fmt.Print(a...)
}

// Printf formats according to a format specifier and writes to standard output.
func Printf(format string, a ...any) {
	fmt.Printf(format, a...)
}

// Println formats using the default formats for its operands and writes to standard output.
func Println(a ...any) {
	fmt.Println(a...)
}

// Sprint formats using the default formats for its operands and returns the resulting string.
func Sprint(a ...any) string {
	return fmt.Sprint(a...)
}

// Sprintf formats according to a format specifier and returns the resulting string.
func Sprintf(format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}

// Sprintln formats using the default formats for its operands and returns the resulting string.
func Sprintln(a ...any) string {
	return fmt.Sprintln(a...)
}

// Fprint formats using the default formats for its operands and writes to w.
func Fprint(w io.Writer, a ...any) (n int, err error) {
	return fmt.Fprint(w, a...)
}

// Fprintf formats according to a format specifier and writes to w.
func Fprintf(w io.Writer, format string, a ...any) (n int, err error) {
	return fmt.Fprintf(w, format, a...)
}

// Fprintln formats using the default formats for its operands and writes to w.
func Fprintln(w io.Writer, a ...any) (n int, err error) {
	return fmt.Fprintln(w, a...)
}

// fmt.Println("\033[31m红色文本\033[0m")  // 红色
// fmt.Println("\033[32m绿色文本\033[0m")  // 绿色
// fmt.Println("\033[33m黄色文本\033[0m")  // 黄色
// fmt.Println("\033[34m蓝色文本\033[0m")  // 蓝色
// fmt.Println("\033[35m紫色文本\033[0m")  // 紫色
// fmt.Println("\033[36m青色文本\033[0m")  // 青色

func ColorPrintf(color string, format string, a ...any) {
	switch color {
	case "red":
		format = "\033[31m" + format + "\033[0m"
	case "green":
		format = "\033[32m" + format + "\033[0m"
	case "yellow":
		format = "\033[33m" + format + "\033[0m"
	case "blue":
		format = "\033[34m" + format + "\033[0m"
	case "purple":
		format = "\033[35m" + format + "\033[0m"
	case "cyan":
		format = "\033[36m" + format + "\033[0m"
	default:
		panic(Sprintf("unknown color: %s", color))
	}

	Printf(format, a...)
}
