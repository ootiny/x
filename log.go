// Package x provides a simplified logging interface wrapping logrus
// with support for different logging levels and formats.
package x

import (
	log "github.com/sirupsen/logrus"
)

// Trace logs a message at level Trace on the standard logger.
func Trace(args ...any) {
	log.Trace(args...)
}

// Debug logs a message at level Debug on the standard logger.
func Debug(args ...any) {
	log.Debug(args...)
}

// Info logs a message at level Info on the standard logger.
func Info(args ...any) {
	log.Info(args...)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...any) {
	log.Warn(args...)
}

// Error logs a message at level Error on the standard logger.
func Error(args ...any) {
	log.Error(args...)
}

// Panic logs a message at level Panic on the standard logger.
func Panic(args ...any) {
	log.Panic(args...)
}

// Fatal logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func Fatal(args ...any) {
	log.Fatal(args...)
}

// Tracef logs a message at level Trace on the standard logger.
func Tracef(format string, args ...any) {
	log.Tracef(format, args...)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(format string, args ...any) {
	log.Debugf(format, args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(format string, args ...any) {
	log.Infof(format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...any) {
	log.Warnf(format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(format string, args ...any) {
	log.Errorf(format, args...)
}

// Panicf logs a message at level Panic on the standard logger.
func Panicf(format string, args ...any) {
	log.Panicf(format, args...)
}

// Fatalf logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func Fatalf(format string, args ...any) {
	log.Fatalf(format, args...)
}

// Traceln logs a message at level Trace on the standard logger.
func Traceln(args ...any) {
	log.Traceln(args...)
}

// Debugln logs a message at level Debug on the standard logger.
func Debugln(args ...any) {
	log.Debugln(args...)
}

// Infoln logs a message at level Info on the standard logger.
func Infoln(args ...any) {
	log.Infoln(args...)
}

// Warnln logs a message at level Warn on the standard logger.
func Warnln(args ...any) {
	log.Warnln(args...)
}

// Errorln logs a message at level Error on the standard logger.
func Errorln(args ...any) {
	log.Errorln(args...)
}

// Panicln logs a message at level Panic on the standard logger.
func Panicln(args ...any) {
	log.Panicln(args...)
}

// Fatalln logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func Fatalln(args ...any) {
	log.Fatalln(args...)
}
