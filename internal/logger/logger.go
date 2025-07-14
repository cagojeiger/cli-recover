package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type Logger interface {
	Debug(msg string, attrs ...any)
	Info(msg string, attrs ...any)
	Warn(msg string, attrs ...any)
	Error(msg string, attrs ...any)
	With(attrs ...any) Logger
	WithContext(ctx context.Context) Logger
}

type Config struct {
	Level      string `yaml:"level"`       // debug, info, warn, error
	Format     string `yaml:"format"`      // text, json
	Output     string `yaml:"output"`      // stdout, stderr, file, both
	FilePath   string `yaml:"file_path"`   // log file path when output includes file
	MaxSize    int    `yaml:"max_size"`    // max file size in MB before rotation
	MaxBackups int    `yaml:"max_backups"` // max number of old log files to retain
	MaxAge     int    `yaml:"max_age"`     // max age in days to retain old log files
}

type logger struct {
	slogger *slog.Logger
}

func (l *logger) Debug(msg string, attrs ...any) {
	l.slogger.Debug(msg, attrs...)
}

func (l *logger) Info(msg string, attrs ...any) {
	l.slogger.Info(msg, attrs...)
}

func (l *logger) Warn(msg string, attrs ...any) {
	l.slogger.Warn(msg, attrs...)
}

func (l *logger) Error(msg string, attrs ...any) {
	l.slogger.Error(msg, attrs...)
}

func (l *logger) With(attrs ...any) Logger {
	return &logger{slogger: l.slogger.With(attrs...)}
}

func (l *logger) WithContext(ctx context.Context) Logger {
	// slog doesn't have WithContext method, so we'll just return the same logger
	// In a real implementation, we might want to extract values from context
	return l
}

var (
	defaultLogger Logger
)

func init() {
	cfg := &Config{
		Level:  "info",
		Format: "text",
		Output: "stderr",
	}
	defaultLogger, _ = New(cfg)
}

func Default() Logger {
	return defaultLogger
}

func SetDefault(l Logger) {
	defaultLogger = l
}

func New(cfg *Config) (Logger, error) {
	if cfg == nil {
		cfg = &Config{
			Level:  "info",
			Format: "text",
			Output: "stderr",
		}
	}

	// Parse log level
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Setup handler options
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize time format
			if a.Key == slog.TimeKey && a.Value.Kind() == slog.KindTime {
				t := a.Value.Time()
				a.Value = slog.StringValue(t.Format(time.RFC3339))
			}
			return a
		},
	}

	// Setup writer
	var w io.Writer
	switch cfg.Output {
	case "stdout":
		w = os.Stdout
	case "stderr":
		w = os.Stderr
	case "file":
		if cfg.FilePath == "" {
			cfg.FilePath = "cli-pipe.log"
		}
		// Use rotating file writer if rotation is configured
		if cfg.MaxSize > 0 || cfg.MaxBackups > 0 || cfg.MaxAge > 0 {
			w = NewRotatingFileWriter(cfg.FilePath, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)
		} else {
			// Ensure directory exists
			dir := filepath.Dir(cfg.FilePath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, err
			}
			file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return nil, err
			}
			w = file
		}
	case "both":
		if cfg.FilePath == "" {
			cfg.FilePath = "cli-pipe.log"
		}
		var fileWriter io.Writer
		// Use rotating file writer if rotation is configured
		if cfg.MaxSize > 0 || cfg.MaxBackups > 0 || cfg.MaxAge > 0 {
			fileWriter = NewRotatingFileWriter(cfg.FilePath, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)
		} else {
			// Ensure directory exists
			dir := filepath.Dir(cfg.FilePath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, err
			}
			file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return nil, err
			}
			fileWriter = file
		}
		w = io.MultiWriter(os.Stderr, fileWriter)
	default:
		w = os.Stderr
	}

	// Create handler based on format
	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}

	slogger := slog.New(handler)
	return &logger{slogger: slogger}, nil
}

// Convenience functions for default logger
func Debug(msg string, attrs ...any) {
	defaultLogger.Debug(msg, attrs...)
}

func Info(msg string, attrs ...any) {
	defaultLogger.Info(msg, attrs...)
}

func Warn(msg string, attrs ...any) {
	defaultLogger.Warn(msg, attrs...)
}

func Error(msg string, attrs ...any) {
	defaultLogger.Error(msg, attrs...)
}

func With(attrs ...any) Logger {
	return defaultLogger.With(attrs...)
}

func WithContext(ctx context.Context) Logger {
	return defaultLogger.WithContext(ctx)
}