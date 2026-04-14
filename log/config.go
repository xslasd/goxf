package log

import (
	"fmt"
	"github.com/xslasd/goxf/application"
	"io"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	// Dir 日志输出目录
	Dir string
	// Name 日志文件名称
	Name string
	// Level 日志初始等级
	Level string
	// 异步存储
	Async bool
	// 异步缓冲池大小（字节）
	BufferSize int
	// 刷新间隔（秒）
	FlushInterval int
	// 日志输出文件最大长度，超过改值则截断
	MaxSize int
	// 日志输出文件保存最大天数
	MaxAge int
	// 日志输出文件保存最大日志数量
	MaxBackup int
	// 日志格式配置
	EncoderConfig *zapcore.EncoderConfig
	// 日志处理配置
	Core zapcore.Core
	// 日志初始化字段
	Fields []zap.Field
	// 是否添加调用者信息
	AddCaller  bool
	CallerSkip int
	// 打印堆栈级别
	StackLevel string
	// 是否控制台输出
	enableConsole bool
	confPrefix    string
	confName      string
}

type Option func(*Config)

func ConfPrefix(prefix string) Option {
	return func(o *Config) {
		o.confPrefix = prefix
	}
}

func ConfName(confName string) Option {
	return func(o *Config) {
		o.confName = confName
	}
}

func WithLogFileName(name string) Option {
	return func(o *Config) {
		o.Name = name
	}
}

func WithEnableConsole(enableConsole bool) Option {
	return func(o *Config) {
		o.enableConsole = enableConsole
	}
}

func (config Config) getFileName() string {
	return filepath.Join(config.Dir, config.Name)
}

func DefaultConfig() *Config {
	return &Config{
		Name:          application.GetServiceName() + ".log",
		Dir:           ".",
		Level:         "info",
		MaxSize:       500, // 500M
		MaxAge:        1,   // 1 day
		MaxBackup:     10,  // 10 backup
		CallerSkip:    1,
		AddCaller:     true,
		Async:         true,
		BufferSize:    256 * 1024,
		FlushInterval: 30, // 30s
		EncoderConfig: DefaultZapEncoderConfig(),
		StackLevel:    "dpanic",
		enableConsole: application.GetEnableConsole(),
		confPrefix:    "logger",
		confName:      "default",
	}
}

func DefaultZapEncoderConfig() *zapcore.EncoderConfig {
	return &zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "lv",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     TimeLayoutEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   ShortCallerEncoder,
	}
}

func TimeLayoutEncoderColor(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(color.YellowString(t.Format("2006-01-02 15:04:05.000")))
}

func TimeLayoutEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func TimeEncoderColor(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(color.YellowString(strconv.FormatInt(t.Unix(), 10)))
}

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendInt64(t.Unix())
}

// ShortCallerEncoder serializes a caller in package/file:line format, trimming
// all but the final directory from the full path.
func ShortCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	// TODO: consider using a byte-oriented API to save an allocation.
	enc.AppendString(fmt.Sprintf("%-32s", caller.TrimmedPath()))
}

// ConsoleEncodeLevel ...
func ConsoleEncodeLevel(lv zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var colorize = color.RedString
	switch lv {
	case zapcore.DebugLevel:
		colorize = color.BlueString
	case zapcore.InfoLevel:
		colorize = color.GreenString
	case zapcore.WarnLevel:
		colorize = color.YellowString
	case zapcore.ErrorLevel, zap.PanicLevel, zap.DPanicLevel, zap.FatalLevel:
		colorize = color.RedString
	default:
	}
	enc.AppendString(colorize(lv.CapitalString()))
}

// getRotate get a *lumberjack.Logger
func getRotate(config *Config) io.Writer {
	return &lumberjack.Logger{
		Filename:   config.getFileName(),
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackup,
		MaxAge:     config.MaxAge,
	}
}
