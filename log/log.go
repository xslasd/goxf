package log

import (
	"fmt"
	"github.com/xslasd/goxf/hooks"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

type Field = zap.Field

const (
	// defaultBufferSize sizes the buffer associated with each WriterSync.
	defaultBufferSize = 256 * 1024
	// defaultFlushInterval means the default flush interval
	defaultFlushInterval = 5 * time.Second
)

type Logger struct {
	config  Config
	zlogger *zap.Logger
	sugar   *zap.SugaredLogger
	lv      *zap.AtomicLevel
}

var (
	// String ...
	String = zap.String
	// Any ...
	Any = zap.Any
	// Int64 ...
	Int64 = zap.Int64
	// Int ...
	Int = zap.Int
	// Int32 ...
	Int32 = zap.Int32
	// Uint ...
	Uint = zap.Uint
	// Duration ...
	Duration = zap.Duration
	// Durationp ...
	Durationp = zap.Durationp
	// Object ...
	Object = zap.Object
	// Namespace ...
	Namespace = zap.Namespace
	// Reflect ...
	Reflect = zap.Reflect
	// Skip ...
	Skip = zap.Skip()
	// ByteString ...
	ByteString = zap.ByteString
)

func newLogger(config *Config) *Logger {
	zapOptions := make([]zap.Option, 0)

	var zl = zapcore.DPanicLevel
	_ = (&zl).UnmarshalText([]byte(config.StackLevel))
	zapOptions = append(zapOptions, zap.AddStacktrace(zl))

	if config.AddCaller {
		zapOptions = append(zapOptions, zap.AddCaller(), zap.AddCallerSkip(config.CallerSkip))
	}

	if len(config.Fields) > 0 {
		zapOptions = append(zapOptions, zap.Fields(config.Fields...))
	}

	var ws zapcore.WriteSyncer
	if config.enableConsole {
		ws = os.Stdout
	} else {
		ws = zapcore.AddSync(getRotate(config))
	}
	if config.Async {
		bws := &zapcore.BufferedWriteSyncer{
			WS:            zapcore.AddSync(ws),
			FlushInterval: defaultFlushInterval,
			Size:          defaultBufferSize,
		}
		hooks.Register(hooks.Stage_AfterStop, func() {
			_ = bws.Sync()
			_ = bws.Stop()
		})
		ws = bws
	}

	lv := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	if err := lv.UnmarshalText([]byte(config.Level)); err != nil {
		panic(err)
	}
	if config.Core == nil {
		encoderConfig := *config.EncoderConfig
		config.Core = zapcore.NewCore(
			func() zapcore.Encoder {
				if config.enableConsole {
					return zapcore.NewConsoleEncoder(encoderConfig)
				}
				return zapcore.NewJSONEncoder(encoderConfig)
			}(),
			ws,
			lv,
		)
	}

	zapLogger := zap.New(
		config.Core,
		zapOptions...,
	)
	return &Logger{
		zlogger: zapLogger,
		lv:      &lv,
		config:  *config,
		sugar:   zapLogger.Sugar(),
	}
}

func (l *Logger) Level(lvText string) {
	l.lv.UnmarshalText([]byte(lvText))
}

// IsDebugMode ...
func (l *Logger) IsEnableConsole() bool {
	return l.config.enableConsole
}

func (l *Logger) ConfName() string {
	return l.config.confName
}

func (l *Logger) GetConfLevelKey() string {
	key := l.config.confPrefix + "." + l.config.confName + ".level"
	return key
}

// Debug ...
func (l *Logger) Debug(msg string, fields ...Field) {
	l.zlogger.Debug(msg, fields...)
}

// Debugw ...
func (l *Logger) Debugw(msg string, keysAndValues ...any) {
	l.sugar.Debugw(msg, keysAndValues...)
}

// Debugf ...
func (l *Logger) Debugf(template string, args ...any) {
	l.sugar.Debugw(sprintf(template, args...))
}

// Info ...
func (l *Logger) Info(msg string, fields ...Field) {
	l.zlogger.Info(msg, fields...)
}

// Infow ...
func (l *Logger) Infow(msg string, keysAndValues ...any) {
	l.sugar.Infow(msg, keysAndValues...)
}

// Infof ...
func (l *Logger) Infof(template string, args ...any) {
	l.sugar.Infow(sprintf(template, args...))
}

// Warn ...
func (l *Logger) Warn(msg string, fields ...Field) {
	l.zlogger.Warn(msg, fields...)
}

// Warnw ...
func (l *Logger) Warnw(msg string, keysAndValues ...any) {
	l.sugar.Warnw(msg, keysAndValues...)
}

// Warnf ...
func (l *Logger) Warnf(template string, args ...any) {
	l.sugar.Warn(sprintf(template, args...))
}

// Error ...
func (l *Logger) Error(msg string, fields ...Field) {
	l.zlogger.Error(msg, fields...)
}

// Errorw ...
func (l *Logger) Errorw(msg string, keysAndValues ...any) {
	l.sugar.Errorw(msg, keysAndValues...)
}

// Errorf ...
func (l *Logger) Errorf(template string, args ...any) {
	l.sugar.Error(sprintf(template, args...))
}

// Panic ...
func (l *Logger) Panic(msg string, fields ...Field) {
	l.zlogger.Panic(msg, fields...)
}

// Panicw ...
func (l *Logger) Panicw(msg string, keysAndValues ...any) {
	l.sugar.Panicw(msg, keysAndValues...)
}

// Panicf ...
func (l *Logger) Panicf(template string, args ...any) {
	l.sugar.Panic(sprintf(template, args...))
}

// DPanic ...
func (l *Logger) DPanic(msg string, fields ...Field) {
	l.zlogger.DPanic(msg, fields...)
}

// DPanicw ...
func (l *Logger) DPanicw(msg string, keysAndValues ...any) {
	l.sugar.DPanicw(msg, keysAndValues...)
}

// DPanicf ...
func (l *Logger) DPanicf(template string, args ...any) {
	l.sugar.DPanic(sprintf(template, args...))
}

// Fatal ...
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.zlogger.Fatal(msg, fields...)
}

// Fatalw ...
func (l *Logger) Fatalw(msg string, keysAndValues ...any) {
	l.sugar.Fatalw(msg, keysAndValues...)
}

// Fatalf ...
func (l *Logger) Fatalf(template string, args ...any) {
	l.sugar.Fatal(sprintf(template, args...))
}

// With ...
func (l *Logger) With(fields ...Field) *Logger {
	desugarLogger := l.zlogger.With(fields...)
	return &Logger{
		zlogger: desugarLogger,
		lv:      l.lv,
		sugar:   desugarLogger.Sugar(),
		config:  l.config,
	}
}

func sprintf(template string, args ...any) string {
	msg := template
	if msg == "" && len(args) > 0 {
		msg = fmt.Sprint(args...)
	} else if msg != "" && len(args) > 0 {
		msg = fmt.Sprintf(template, args...)
	}
	return msg
}
