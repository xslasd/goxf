package log

import (
	"fmt"
)

// loggerWrapper wraps Logger into a LoggerV2.
type loggerWrapper struct {
	logger *Logger
}

// Info logs to INFO log
func (l *loggerWrapper) Info(args ...any) {
	l.logger.Info(sprint(args...))
}

// Infoln logs to INFO log
func (l *loggerWrapper) Infoln(args ...any) {
	l.logger.Info(sprint(args...))
}

// Infof logs to INFO log
func (l *loggerWrapper) Infof(format string, args ...any) {
	l.logger.Infof(format, args...)
}

// Warning logs to WARNING log
func (l *loggerWrapper) Warning(args ...any) {
	l.logger.Warn(sprint(args...))
}

// Warning logs to WARNING log
func (l *loggerWrapper) Warningln(args ...any) {
	l.logger.Warn(sprint(args...))
}

// Warning logs to WARNING log
func (l *loggerWrapper) Warningf(format string, args ...any) {
	l.logger.Warnf(format, args...)
}

// Error logs to ERROR log
func (l *loggerWrapper) Error(args ...any) {
	l.logger.Error(sprint(args...))
}

// Errorn logs to ERROR log
func (l *loggerWrapper) Errorln(args ...any) {
	l.logger.Error(sprint(args...))
}

// Errorf logs to ERROR log
func (l *loggerWrapper) Errorf(format string, args ...any) {
	l.logger.Errorf(format, args...)
}

// Fatal logs to ERROR log
func (l *loggerWrapper) Fatal(args ...any) {
	l.logger.Fatal(sprint(args...))
}

// Fatalln logs to ERROR log
func (l *loggerWrapper) Fatalln(args ...any) {
	l.logger.Fatal(sprint(args...))
}

// Error logs to ERROR log
func (l *loggerWrapper) Fatalf(format string, args ...any) {
	l.logger.Fatalf(format, args...)
}

// v returns true for all verbose level.
func (l *loggerWrapper) V(v int) bool {
	return true
}

func sprint(args ...any) string {
	return fmt.Sprint(args...)
}
