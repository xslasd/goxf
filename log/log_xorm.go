package log

import (
	"fmt"
	"go.uber.org/zap/zapcore"
	"xorm.io/xorm/log"
)

type xormLoggerWrapper struct {
	logger  *Logger
	showSQL bool
}

func (x xormLoggerWrapper) BeforeSQL(ctx log.LogContext) {
}

func (x xormLoggerWrapper) AfterSQL(ctx log.LogContext) {
	var sessionPart string
	v := ctx.Ctx.Value(log.SessionIDKey)
	if key, ok := v.(string); ok {
		sessionPart = fmt.Sprintf(" [%s]", key)
	}
	if ctx.ExecuteTime > 0 {
		x.logger.Infof("[SQL]%s %s %v - %v", sessionPart, ctx.SQL, ctx.Args, ctx.ExecuteTime)
	} else {
		x.logger.Infof("[SQL]%s %s %v", sessionPart, ctx.SQL, ctx.Args)
	}
}

func (x xormLoggerWrapper) Debugf(format string, v ...any) {
	x.logger.Debugf(format, v...)
}

func (x xormLoggerWrapper) Errorf(format string, v ...any) {
	x.logger.Errorf(format, v...)
}

func (x xormLoggerWrapper) Infof(format string, v ...any) {
	x.logger.Infof(format, v...)
}

func (x xormLoggerWrapper) Warnf(format string, v ...any) {
	x.logger.Warnf(format, v...)
}

func (x xormLoggerWrapper) Level() log.LogLevel {
	switch x.logger.lv.Level() {
	case zapcore.DebugLevel:
		return log.LOG_DEBUG
	case zapcore.InfoLevel:
		return log.LOG_INFO
	case zapcore.WarnLevel:
		return log.LOG_WARNING
	case zapcore.ErrorLevel:
		return log.LOG_ERR
	default:
		return log.LOG_UNKNOWN
	}
}

func (x xormLoggerWrapper) SetLevel(l log.LogLevel) {
}

func (x *xormLoggerWrapper) ShowSQL(show ...bool) {
	if len(show) == 0 {
		x.showSQL = true
		return
	}
	x.showSQL = show[0]
}

func (x xormLoggerWrapper) IsShowSQL() bool {
	return x.showSQL
}
