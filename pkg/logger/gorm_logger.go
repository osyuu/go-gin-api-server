package logger

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

// GormLogger 自定義GORM logger使用我們的zap logger
type GormLogger struct {
	zapLogger *zap.Logger
}

// NewGormLogger 創建新的GORM logger實例
func NewGormLogger() logger.Interface {
	return &GormLogger{zapLogger: Log}
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.zapLogger.Info(fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.zapLogger.Warn(fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.zapLogger.Error(fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if err != nil {
		sql, _ := fc()
		l.zapLogger.Error("SQL Error",
			zap.String("sql", sql),
			zap.Duration("duration", time.Since(begin)),
			zap.Error(err))
	} else {
		sql, rows := fc()
		l.zapLogger.Debug("SQL Query",
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("duration", time.Since(begin)))
	}
}
