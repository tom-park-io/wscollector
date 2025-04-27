package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"wscollector/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// New creates a zap.Logger configured based on the given options.
func New(opts config.LogConfig) (*zap.Logger, error) {
	// Parse log level
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(opts.Level)); err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	// Determine encoding format
	encoding := "json"
	if opts.Environment == "dev" || opts.Format == "console" {
		encoding = "console"
	}

	// Get encoder config based on format
	encoderCfg := encoderConfig(encoding)

	// Compose output cores
	var cores []zapcore.Core

	// Console (stdout) output
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		lvl,
	)
	cores = append(cores, consoleCore)

	// Optional file output with rotation via lumberjack
	if opts.OutputFile != "" {
		if opts.OutputFile != "" {
			// Create parent directory if it doesn't exist
			dir := filepath.Dir(opts.OutputFile)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create log directory: %w", err)
			}
		}

		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   opts.OutputFile,
			MaxSize:    10,   // max file size (MB) before rotation
			MaxBackups: 5,    // max number of old log files to keep
			MaxAge:     7,    // max age (days) to retain a log file
			Compress:   true, // compress rotated files
		})

		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			fileWriter,
			lvl,
		)
		cores = append(cores, fileCore)
	}

	// Combine all cores
	core := zapcore.NewTee(cores...)

	// Build the logger with caller and stacktrace options
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return logger, nil
}

// encoderConfig returns a zapcore.EncoderConfig based on log format.
func encoderConfig(format string) zapcore.EncoderConfig {
	if format == "console" {
		return zap.NewDevelopmentEncoderConfig()
	}
	return zap.NewProductionEncoderConfig()
}
