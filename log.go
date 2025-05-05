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

type LogCallerDebugConfig struct {
	TraceMaxDepth uint32
	DebugMaxDepth uint32
	InfoMaxDepth  uint32
	WarnMaxDepth  uint32
	ErrorMaxDepth uint32
	FatalMaxDepth uint32
	PanicMaxDepth uint32
}

var (
	maxCallerDepthConfig = LogCallerDebugConfig{}
)

func SetLogLevel(level log.Level) {
	log.SetLevel(level)
}

func SetCallerDebugConfig(config LogCallerDebugConfig) {
	maxCallerDepthConfig = config
}

func ReportCaller(maxDepth uint32) {
	if maxDepth == 0 {
		return
	}
	sb := bytes.Buffer{}
	pcs := make([]uintptr, 100)
	n := runtime.Callers(0, pcs)

	for i := 2; i < n && maxDepth > 0; i++ {
		maxDepth--

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
	ReportCaller(maxCallerDepthConfig.TraceMaxDepth)
}

// LogDebug logs a message at level Debug on the standard logger.
func LogDebug(args ...any) {
	log.Debug(args...)
	ReportCaller(maxCallerDepthConfig.DebugMaxDepth)
}

// LogInfo logs a message at level Info on the standard logger.
func LogInfo(args ...any) {
	log.Info(args...)
	ReportCaller(maxCallerDepthConfig.InfoMaxDepth)
}

// LogWarn logs a message at level Warn on the standard logger.
func LogWarn(args ...any) {
	log.Warn(args...)
	ReportCaller(maxCallerDepthConfig.WarnMaxDepth)
}

// LogError logs a message at level Error on the standard logger.
func LogError(args ...any) {
	log.Error(args...)
	ReportCaller(maxCallerDepthConfig.ErrorMaxDepth)
}

// LogPanic logs a message at level Panic on the standard logger.
func LogPanic(args ...any) {
	log.Panic(args...)
	ReportCaller(maxCallerDepthConfig.PanicMaxDepth)
}

// LogFatal logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func LogFatal(args ...any) {
	log.Fatal(args...)
	ReportCaller(maxCallerDepthConfig.FatalMaxDepth)
}

// LogTracef logs a message at level Trace on the standard logger.
func LogTracef(format string, args ...any) {
	log.Tracef(format, args...)
	ReportCaller(maxCallerDepthConfig.TraceMaxDepth)
}

// LogDebugf logs a message at level Debug on the standard logger.
func LogDebugf(format string, args ...any) {
	log.Debugf(format, args...)
	ReportCaller(maxCallerDepthConfig.DebugMaxDepth)
}

// LogInfof logs a message at level Info on the standard logger.
func LogInfof(format string, args ...any) {
	log.Infof(format, args...)
	ReportCaller(maxCallerDepthConfig.InfoMaxDepth)
}

// LogWarnf logs a message at level Warn on the standard logger.
func LogWarnf(format string, args ...any) {
	log.Warnf(format, args...)
	ReportCaller(maxCallerDepthConfig.WarnMaxDepth)
}

// LogErrorf logs a message at level Error on the standard logger.
func LogErrorf(format string, args ...any) {
	log.Errorf(format, args...)
	ReportCaller(maxCallerDepthConfig.ErrorMaxDepth)
}

// LogPanicf logs a message at level Panic on the standard logger.
func LogPanicf(format string, args ...any) {
	log.Panicf(format, args...)
	ReportCaller(maxCallerDepthConfig.PanicMaxDepth)
}

// LogFatalf logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func LogFatalf(format string, args ...any) {
	log.Fatalf(format, args...)
	ReportCaller(maxCallerDepthConfig.FatalMaxDepth)
}

// LogTraceln logs a message at level Trace on the standard logger.
func LogTraceln(args ...any) {
	log.Traceln(args...)
	ReportCaller(maxCallerDepthConfig.TraceMaxDepth)
}

// LogDebugln logs a message at level Debug on the standard logger.
func LogDebugln(args ...any) {
	log.Debugln(args...)
	ReportCaller(maxCallerDepthConfig.DebugMaxDepth)
}

// LogInfoln logs a message at level Info on the standard logger.
func LogInfoln(args ...any) {
	log.Infoln(args...)
	ReportCaller(maxCallerDepthConfig.InfoMaxDepth)
}

// LogWarnln logs a message at level Warn on the standard logger.
func LogWarnln(args ...any) {
	log.Warnln(args...)
	ReportCaller(maxCallerDepthConfig.WarnMaxDepth)
}

// LogErrorln logs a message at level Error on the standard logger.
func LogErrorln(args ...any) {
	log.Errorln(args...)
	ReportCaller(maxCallerDepthConfig.ErrorMaxDepth)
}

// LogPanicln logs a message at level Panic on the standard logger.
func LogPanicln(args ...any) {
	log.Panicln(args...)
	ReportCaller(maxCallerDepthConfig.PanicMaxDepth)
}

// LogFatalln logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func LogFatalln(args ...any) {
	log.Fatalln(args...)
	ReportCaller(maxCallerDepthConfig.FatalMaxDepth)
}
