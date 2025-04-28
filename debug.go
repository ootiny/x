package x

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
)

func DebugTracef(fnDeepth uint, format string, args ...any) string {
	sb := bytes.Buffer{}
	header := fmt.Sprintf(format, args...)

	if _, file, line, ok := runtime.Caller(int(fnDeepth) + 1); ok && line > 0 {
		if header != "" {
			sb.WriteString(header)
			sb.WriteString(" ")
		}

		sb.WriteString(file)
		sb.WriteByte(':')
		sb.WriteString(strconv.Itoa(line))
	} else {
		sb.WriteString(header)
	}

	return sb.String()
}
