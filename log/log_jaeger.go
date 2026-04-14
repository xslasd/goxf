package log

type jaegerLoggerWrapper struct {
	logger *Logger
}

func (j jaegerLoggerWrapper) Error(msg string) {
	j.logger.Error(msg)
}

func (j jaegerLoggerWrapper) Infof(msg string, args ...any) {
	j.logger.Debugf(msg, args...)
}
