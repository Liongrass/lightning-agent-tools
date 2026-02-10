// Copyright (c) 2025 Lightning Labs
// Distributed under the MIT license. See LICENSE for details.

// Package logging provides structured logging functionality for the MCP LNC server.
// It builds on zap for high-performance logging in both development and
// production configurations.
package logging

import (
	"os"

	"github.com/lightninglabs/lightning-agent-kit/lightning-mcp-server/internal/interfaces"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the global logger instance.
var Logger *zap.Logger

// ZapLogger wraps zap.Logger to implement interfaces.Logger.
type zapLogger struct {
	logger *zap.Logger
}

// NewLogger creates a new logger that implements interfaces.Logger.
func NewLogger(logger *zap.Logger) interfaces.Logger {
	return &zapLogger{logger: logger}
}

// Debug logs a debug message.
func (l *zapLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

// Info logs an info message.
func (l *zapLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// Warn logs a warning message.
func (l *zapLogger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

// Error logs an error message.
func (l *zapLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

// Fatal logs a fatal message and exits.
func (l *zapLogger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

// With creates a child logger with additional fields.
func (l *zapLogger) With(fields ...zap.Field) interfaces.Logger {
	return &zapLogger{logger: l.logger.With(fields...)}
}

// InitLogger initializes the global logger with appropriate configuration.
func InitLogger(development bool) error {
	var config zap.Config

	if development {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}

	// Always log to stderr for MCP compatibility (stdout is used for MCP protocol)
	config.OutputPaths = []string{"stderr"}
	config.ErrorOutputPaths = []string{"stderr"}

	// Set log level based on environment variable
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		var zapLevel zapcore.Level
		if err := zapLevel.UnmarshalText([]byte(level)); err == nil {
			config.Level.SetLevel(zapLevel)
		}
	}

	logger, err := config.Build()
	if err != nil {
		return err
	}

	Logger = logger
	zap.ReplaceGlobals(logger)

	return nil
}

// Sync flushes any buffered log entries.
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
