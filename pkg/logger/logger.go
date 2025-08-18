// Package logger provides a global zap logger instance for JSON formatted logging
// across the entire langfuse-go project.
package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.Logger
	once         sync.Once
)

// Config holds the logger configuration options
type Config struct {
	Level      zapcore.Level
	Encoding   string
	OutputPath string
	ErrorPath  string
}

// DefaultConfig returns a default production-ready logger configuration
func DefaultConfig() Config {
	return Config{
		Level:      zapcore.InfoLevel,
		Encoding:   "json",
		OutputPath: "stdout",
		ErrorPath:  "stderr",
	}
}

// Init initializes the global logger with the provided configuration.
// It's safe to call multiple times - only the first call will initialize the logger.
func Init(config Config) error {
	var err error
	once.Do(func() {
		zapConfig := zap.NewProductionConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(config.Level)
		zapConfig.Encoding = config.Encoding
		zapConfig.OutputPaths = []string{config.OutputPath}
		zapConfig.ErrorOutputPaths = []string{config.ErrorPath}

		// Customize the encoder config for better JSON output
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.LevelKey = "level"
		zapConfig.EncoderConfig.MessageKey = "message"
		zapConfig.EncoderConfig.CallerKey = "caller"
		zapConfig.EncoderConfig.StacktraceKey = "stacktrace"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		zapConfig.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		zapConfig.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

		globalLogger, err = zapConfig.Build(zap.AddCallerSkip(1))
		if err != nil {
			// Fallback to no-op logger if configuration fails
			globalLogger = zap.NewNop()
		}
	})
	return err
}

// Get returns the global logger instance.
// If the logger hasn't been initialized, it will be initialized with default config.
func Get() *zap.Logger {
	if globalLogger == nil {
		_ = Init(DefaultConfig())
	}
	return globalLogger
}
