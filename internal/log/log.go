// Package log provides a two-layer logger:
//   - Developer layer: structured JSON to a rolling file (full details)
//   - User layer: plain messages surfaced through the gRPC log stream
package log

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/lumberjack.v2"
)

// Logger wraps a zap.Logger and exposes a channel for user-facing log entries.
type Logger struct {
	zap     *zap.Logger
	entries chan Entry
}

// Entry is a user-facing log entry pushed to the gRPC WatchLogs stream.
type Entry struct {
	Level   Level
	Message string
	Fields  map[string]string
}

type Level int32

const (
	LevelDebug Level = 1
	LevelInfo  Level = 2
	LevelWarn  Level = 3
	LevelError Level = 4
)

// New creates a Logger that writes structured JSON to logDir/weave.log (rolling)
// and to stderr in development mode.
func New(logDir string, debug bool) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0o750); err != nil {
		return nil, err
	}

	roller := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "weave.log"),
		MaxSize:    50, // MB
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	}

	level := zapcore.InfoLevel
	if debug {
		level = zapcore.DebugLevel
	}

	fileEnc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	fileCore := zapcore.NewCore(fileEnc, zapcore.AddSync(roller), level)

	cores := []zapcore.Core{fileCore}
	if debug {
		consoleEnc := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		consoleCore := zapcore.NewCore(consoleEnc, zapcore.AddSync(os.Stderr), zapcore.DebugLevel)
		cores = append(cores, consoleCore)
	}

	z := zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{
		zap:     z,
		entries: make(chan Entry, 256),
	}, nil
}

// Nop returns a no-op logger (useful in tests).
func Nop() *Logger {
	return &Logger{zap: zap.NewNop(), entries: make(chan Entry, 1)}
}

func (l *Logger) Debug(msg string, fields ...zap.Field) { l.zap.Debug(msg, fields...) }
func (l *Logger) Info(msg string, fields ...zap.Field)  { l.zap.Info(msg, fields...); l.emit(LevelInfo, msg) }
func (l *Logger) Warn(msg string, fields ...zap.Field)  { l.zap.Warn(msg, fields...); l.emit(LevelWarn, msg) }
func (l *Logger) Error(msg string, fields ...zap.Field) { l.zap.Error(msg, fields...); l.emit(LevelError, msg) }
func (l *Logger) Fatal(msg string, fields ...zap.Field) { l.zap.Fatal(msg, fields...) }

// With returns a child logger with pre-set fields.
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{zap: l.zap.With(fields...), entries: l.entries}
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() { _ = l.zap.Sync() }

// Entries returns a read-only channel of user-facing log entries.
// The gRPC WatchLogs handler reads from this channel.
func (l *Logger) Entries() <-chan Entry { return l.entries }

// WriterAt returns an io.Writer that logs each line at Info level.
// Used to capture sing-box's internal output.
func (l *Logger) WriterAt(level Level) io.Writer {
	return &lineWriter{log: l, level: level}
}

func (l *Logger) emit(level Level, msg string) {
	select {
	case l.entries <- Entry{Level: level, Message: msg}:
	default: // drop if buffer full; never block the caller
	}
}

// WithContext is a convenience for passing the logger via context.
type contextKey struct{}

func WithContext(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(contextKey{}).(*Logger); ok {
		return l
	}
	return Nop()
}

// lineWriter implements io.Writer, splitting on newlines and logging each line.
type lineWriter struct {
	log   *Logger
	level Level
	buf   []byte
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)
	for {
		idx := -1
		for i, b := range w.buf {
			if b == '\n' {
				idx = i
				break
			}
		}
		if idx < 0 {
			break
		}
		line := string(w.buf[:idx])
		w.buf = w.buf[idx+1:]
		switch w.level {
		case LevelDebug:
			w.log.Debug(line)
		case LevelWarn:
			w.log.Warn(line)
		case LevelError:
			w.log.Error(line)
		default:
			w.log.Info(line)
		}
	}
	return len(p), nil
}
