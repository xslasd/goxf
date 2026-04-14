package log

import (
	"github.com/pkg/errors"
	"google.golang.org/grpc/grpclog"
)

var (
	defaultLogger      *Logger
	ConfigHandleFun    ConfigHandle
	DefaultLoggerIsNil = errors.New("the default logger is nil")
)

type ConfigHandle func(key string, c *Config) error

func SetLoggerConfigHandle(fn ConfigHandle) {
	ConfigHandleFun = fn
}

func NewLogger(opts ...Option) *Logger {
	config := DefaultConfig()
	for _, o := range opts {
		o(config)
	}
	key := config.confPrefix + "." + config.confName
	if ConfigHandleFun != nil {
		if err := ConfigHandleFun(key, config); err != nil {
			panic(err)
		}
	}
	if config.EncoderConfig == nil {
		config.EncoderConfig = DefaultZapEncoderConfig()
	}
	if config.enableConsole {
		config.EncoderConfig.EncodeLevel = ConsoleEncodeLevel
		config.EncoderConfig.EncodeTime = TimeLayoutEncoderColor
	}
	logger := newLogger(config)
	return logger
}

// SetLogger sets the defaultLogger
func SetLogger(logger *Logger) {
	defaultLogger = logger
}

// SetGRPCLogger sets grpclog's logger using defaultLogger
func SetGRPCLogger() error {
	if defaultLogger != nil {
		grpclog.SetLoggerV2(&loggerWrapper{logger: defaultLogger})
		return nil
	}
	return DefaultLoggerIsNil
}

func GetXormLogger(showSQL bool) (*xormLoggerWrapper, error) {
	if defaultLogger != nil {
		return &xormLoggerWrapper{logger: defaultLogger, showSQL: showSQL}, nil
	}
	return nil, DefaultLoggerIsNil
}

func GetJaegerLogger() (*jaegerLoggerWrapper, error) {
	if defaultLogger != nil {
		return &jaegerLoggerWrapper{logger: defaultLogger}, nil
	}
	return nil, DefaultLoggerIsNil
}

// Info ...
func Info(msg string, fields ...Field) {
	defaultLogger.Info(msg, fields...)
}

// Debug ...
func Debug(msg string, fields ...Field) {
	defaultLogger.Debug(msg, fields...)
}

// Warn ...
func Warn(msg string, fields ...Field) {
	defaultLogger.Warn(msg, fields...)
}

// Error ...
func Error(msg string, fields ...Field) {
	defaultLogger.Error(msg, fields...)
}

// Panic ...
func Panic(msg string, fields ...Field) {
	defaultLogger.Panic(msg, fields...)
}

// DPanic ...
func DPanic(msg string, fields ...Field) {
	defaultLogger.DPanic(msg, fields...)
}

// Fatal ...
func Fatal(msg string, fields ...Field) {
	defaultLogger.Fatal(msg, fields...)
}

// Debugw ...
func Debugw(msg string, keysAndValues ...any) {
	defaultLogger.Debugw(msg, keysAndValues...)
}

// Infow ...
func Infow(msg string, keysAndValues ...any) {
	defaultLogger.Infow(msg, keysAndValues...)
}

// Warnw ...
func Warnw(msg string, keysAndValues ...any) {
	defaultLogger.Warnw(msg, keysAndValues...)
}

// Errorw ...
func Errorw(msg string, keysAndValues ...any) {
	defaultLogger.Errorw(msg, keysAndValues...)
}

// Panicw ...
func Panicw(msg string, keysAndValues ...any) {
	defaultLogger.Panicw(msg, keysAndValues...)
}

// DPanicw ...
func DPanicw(msg string, keysAndValues ...any) {
	defaultLogger.DPanicw(msg, keysAndValues...)
}

// Fatalw ...
func Fatalw(msg string, keysAndValues ...any) {
	defaultLogger.Fatalw(msg, keysAndValues...)
}

// Debugf ...
func Debugf(msg string, args ...any) {
	defaultLogger.Debugf(msg, args...)
}

// Infof ...
func Infof(msg string, args ...any) {
	defaultLogger.Infof(msg, args...)
}

// Warnf ...
func Warnf(msg string, args ...any) {
	defaultLogger.Warnf(msg, args...)
}

// Errorf ...
func Errorf(msg string, args ...any) {
	defaultLogger.Errorf(msg, args...)
}

// Panicf ...
func Panicf(msg string, args ...any) {
	defaultLogger.Panicf(msg, args...)
}

// DPanicf ...
func DPanicf(msg string, args ...any) {
	defaultLogger.DPanicf(msg, args...)
}

// Fatalf ...
func Fatalf(msg string, args ...any) {
	defaultLogger.Fatalf(msg, args...)
}

// With ...
func With(fields ...Field) *Logger {
	return defaultLogger.With(fields...)
}
