// Package x provides a simplified logging interface wrapping logrus
// with support for different logging levels and formats.
package x

import (
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

func SetLogLevel(level log.Level) {
	log.SetLevel(level)
}

// LogTrace logs a message at level Trace on the standard logger.
func LogTrace(args ...any) {
	log.Trace(args...)
}

// LogDebug logs a message at level Debug on the standard logger.
func LogDebug(args ...any) {
	log.Debug(args...)
}

// LogInfo logs a message at level Info on the standard logger.
func LogInfo(args ...any) {
	log.Info(args...)
}

// LogWarn logs a message at level Warn on the standard logger.
func LogWarn(args ...any) {
	log.Warn(args...)
}

// LogError logs a message at level Error on the standard logger.
func LogError(args ...any) {
	log.Error(args...)
}

// LogPanic logs a message at level Panic on the standard logger.
func LogPanic(args ...any) {
	log.Panic(args...)
}

// LogFatal logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func LogFatal(args ...any) {
	log.Fatal(args...)
}

// LogTracef logs a message at level Trace on the standard logger.
func LogTracef(format string, args ...any) {
	log.Tracef(format, args...)
}

// LogDebugf logs a message at level Debug on the standard logger.
func LogDebugf(format string, args ...any) {
	log.Debugf(format, args...)
}

// LogInfof logs a message at level Info on the standard logger.
func LogInfof(format string, args ...any) {
	log.Infof(format, args...)
}

// LogWarnf logs a message at level Warn on the standard logger.
func LogWarnf(format string, args ...any) {
	log.Warnf(format, args...)
}

// LogErrorf logs a message at level Error on the standard logger.
func LogErrorf(format string, args ...any) {
	log.Errorf(format, args...)
}

// LogPanicf logs a message at level Panic on the standard logger.
func LogPanicf(format string, args ...any) {
	log.Panicf(format, args...)
}

// LogFatalf logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func LogFatalf(format string, args ...any) {
	log.Fatalf(format, args...)
}

// LogTraceln logs a message at level Trace on the standard logger.
func LogTraceln(args ...any) {
	log.Traceln(args...)
}

// LogDebugln logs a message at level Debug on the standard logger.
func LogDebugln(args ...any) {
	log.Debugln(args...)
}

// LogInfoln logs a message at level Info on the standard logger.
func LogInfoln(args ...any) {
	log.Infoln(args...)
}

// LogWarnln logs a message at level Warn on the standard logger.
func LogWarnln(args ...any) {
	log.Warnln(args...)
}

// LogErrorln logs a message at level Error on the standard logger.
func LogErrorln(args ...any) {
	log.Errorln(args...)
}

// LogPanicln logs a message at level Panic on the standard logger.
func LogPanicln(args ...any) {
	log.Panicln(args...)
}

// LogFatalln logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func LogFatalln(args ...any) {
	log.Fatalln(args...)
}
