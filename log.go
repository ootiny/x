// Package x provides a simplified logging interface wrapping logrus
// with support for different logging levels and formats.
package x

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const (
	PanicLevel = log.PanicLevel
	FatalLevel = log.FatalLevel
	ErrorLevel = log.ErrorLevel
	WarnLevel  = log.WarnLevel
	InfoLevel  = log.InfoLevel
	DebugLevel = log.DebugLevel
	TraceLevel = log.TraceLevel
)

var (
	maxReportCallerDepth = uint32(0)
)

func SetLogLevel(level log.Level) {
	log.SetLevel(level)
}

func SetMaxReportCallerDepth(n uint32) {
	maxReportCallerDepth = n
}

func ReportCaller() {
	if maxReportCallerDepth == 0 {
		return
	}
	sb := bytes.Buffer{}
	pcs := make([]uintptr, 100)
	n := runtime.Callers(0, pcs)

	needReports := int(maxReportCallerDepth)

	for i := 2; i < n && needReports > 0; i++ {
		needReports--

		if _, file, line, ok := runtime.Caller(i); ok && line > 0 {
			sb.WriteByte('\t')
			sb.WriteString(file)
			sb.WriteByte(':')
			sb.WriteString(strconv.Itoa(line))
			sb.WriteByte('\r')
			sb.WriteByte('\n')
		}
	}

	fmt.Print(sb.String())
}

// LogTrace logs a message at level Trace on the standard logger.
func LogTrace(args ...any) {
	log.Trace(args...)
	ReportCaller()
}

// LogDebug logs a message at level Debug on the standard logger.
func LogDebug(args ...any) {
	log.Debug(args...)
	ReportCaller()
}

// LogInfo logs a message at level Info on the standard logger.
func LogInfo(args ...any) {
	log.Info(args...)
	ReportCaller()
}

// LogWarn logs a message at level Warn on the standard logger.
func LogWarn(args ...any) {
	log.Warn(args...)
	ReportCaller()
}

// LogError logs a message at level Error on the standard logger.
func LogError(args ...any) {
	log.Error(args...)
	ReportCaller()
}

// LogPanic logs a message at level Panic on the standard logger.
func LogPanic(args ...any) {
	log.Panic(args...)
	ReportCaller()
}

// LogFatal logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func LogFatal(args ...any) {
	log.Fatal(args...)
	ReportCaller()
}

// LogTracef logs a message at level Trace on the standard logger.
func LogTracef(format string, args ...any) {
	log.Tracef(format, args...)
	ReportCaller()
}

// LogDebugf logs a message at level Debug on the standard logger.
func LogDebugf(format string, args ...any) {
	log.Debugf(format, args...)
	ReportCaller()
}

// LogInfof logs a message at level Info on the standard logger.
func LogInfof(format string, args ...any) {
	log.Infof(format, args...)
	ReportCaller()
}

// LogWarnf logs a message at level Warn on the standard logger.
func LogWarnf(format string, args ...any) {
	log.Warnf(format, args...)
	ReportCaller()
}

// LogErrorf logs a message at level Error on the standard logger.
func LogErrorf(format string, args ...any) {
	log.Errorf(format, args...)
	ReportCaller()
}

// LogPanicf logs a message at level Panic on the standard logger.
func LogPanicf(format string, args ...any) {
	log.Panicf(format, args...)
	ReportCaller()
}

// LogFatalf logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func LogFatalf(format string, args ...any) {
	log.Fatalf(format, args...)
	ReportCaller()
}

// LogTraceln logs a message at level Trace on the standard logger.
func LogTraceln(args ...any) {
	log.Traceln(args...)
	ReportCaller()
}

// LogDebugln logs a message at level Debug on the standard logger.
func LogDebugln(args ...any) {
	log.Debugln(args...)
	ReportCaller()
}

// LogInfoln logs a message at level Info on the standard logger.
func LogInfoln(args ...any) {
	log.Infoln(args...)
	ReportCaller()
}

// LogWarnln logs a message at level Warn on the standard logger.
func LogWarnln(args ...any) {
	log.Warnln(args...)
	ReportCaller()
}

// LogErrorln logs a message at level Error on the standard logger.
func LogErrorln(args ...any) {
	log.Errorln(args...)
	ReportCaller()
}

// LogPanicln logs a message at level Panic on the standard logger.
func LogPanicln(args ...any) {
	log.Panicln(args...)
	ReportCaller()
}

// LogFatalln logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func LogFatalln(args ...any) {
	log.Fatalln(args...)
	ReportCaller()
}
