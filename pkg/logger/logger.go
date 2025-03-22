package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

type Level uint32

const (
	LevelDebug = Level(0)
	LevelInfo  = Level(1)
	LevelWarn  = Level(2)
	LevelError = Level(3)
)

func (x Level) String() string {
	switch x {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", uint32(x))
	}
}

type ILogger interface {
	Logf(ctx context.Context, level Level, message string, args ...interface{})
}

type DefaultLogger struct {
	logger   *slog.Logger
	logLevel Level
}

func NewDefaultLogger(module string, logLevel Level) (*DefaultLogger, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	logDIR := filepath.Join(home, ".mcp4go")
	_ = os.MkdirAll(logDIR, 0o755)
	fw, err := os.OpenFile(filepath.Join(logDIR, module+".log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	return &DefaultLogger{
		logger:   slog.New(slog.NewTextHandler(fw, nil)),
		logLevel: logLevel,
	}, nil
}

func (x *DefaultLogger) Logf(_ context.Context, level Level, message string, args ...interface{}) {
	if level < x.logLevel {
		return
	}
	// slog print
	switch level {
	case LevelDebug:
		x.logger.Debug(fmt.Sprintf(message, args...))
	case LevelInfo:
		x.logger.Info(fmt.Sprintf(message, args...))
	case LevelWarn:
		x.logger.Warn(fmt.Sprintf(message, args...))
	case LevelError:
		x.logger.Error(fmt.Sprintf(message, args...))
	}
}

type LogHelper struct {
	logger ILogger
}

func NewLogHelper(logger ILogger) *LogHelper {
	return &LogHelper{
		logger: logger,
	}
}

func (x *LogHelper) Debugf(ctx context.Context, message string, args ...interface{}) {
	x.logger.Logf(ctx, LevelDebug, message, args...)
}

func (x *LogHelper) Infof(ctx context.Context, message string, args ...interface{}) {
	x.logger.Logf(ctx, LevelInfo, message, args...)
}

func (x *LogHelper) Warnf(ctx context.Context, message string, args ...interface{}) {
	x.logger.Logf(ctx, LevelWarn, message, args...)
}

func (x *LogHelper) Errorf(ctx context.Context, message string, args ...interface{}) {
	x.logger.Logf(ctx, LevelError, message, args...)
}
