// Copyright (c) 2025 Lightning Labs
// Distributed under the MIT license. See LICENSE for details.

// Package logging provides context-aware logging functionality.
package logging

import (
	"context"

	lnccontext "github.com/lightninglabs/lightning-agent-kit/mcp-server/internal/context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ContextLogger wraps zap.Logger with context awareness.
type ContextLogger struct {
	logger *zap.Logger
}

// NewContextLogger creates a new context-aware logger.
func NewContextLogger(logger *zap.Logger) *ContextLogger {
	if logger == nil {
		logger = zap.L()
	}
	return &ContextLogger{logger: logger}
}

// WithContext creates a logger with context fields automatically included.
func (cl *ContextLogger) WithContext(ctx context.Context) *zap.Logger {
	fields := cl.extractContextFields(ctx)
	if len(fields) == 0 {
		return cl.logger
	}
	return cl.logger.With(fields...)
}

// ExtractContextFields extracts logging fields from context.
func (cl *ContextLogger) extractContextFields(ctx context.Context) []zap.Field {
	var fields []zap.Field

	// Try to get RequestContext first for all fields
	if rc, ok := lnccontext.FromContext(ctx); ok {
		if rc.RequestID() != "" {
			fields = append(fields, zap.String("request_id", rc.RequestID()))
		}
		if rc.TraceID() != "" {
			fields = append(fields, zap.String("trace_id", rc.TraceID()))
		}
		if rc.UserID() != "" {
			fields = append(fields, zap.String("user_id", rc.UserID()))
		}
		if rc.SessionID() != "" {
			fields = append(fields, zap.String("session_id", rc.SessionID()))
		}
		if rc.NodeID() != "" {
			fields = append(fields, zap.String("node_id", rc.NodeID()))
		}
		if rc.Operation() != "" {
			fields = append(fields, zap.String("operation", rc.Operation()))
		}
		fields = append(fields,
			zap.Duration("duration", rc.Duration()),
			zap.Duration("time_remaining", rc.TimeRemaining()),
		)
		return fields
	}

	// Fall back to individual field extraction
	if requestID := lnccontext.GetRequestID(ctx); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}
	if traceID := lnccontext.GetTraceID(ctx); traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}
	if userID := lnccontext.GetUserID(ctx); userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}
	if sessionID := lnccontext.GetSessionID(ctx); sessionID != "" {
		fields = append(fields, zap.String("session_id", sessionID))
	}
	if nodeID := lnccontext.GetNodeID(ctx); nodeID != "" {
		fields = append(fields, zap.String("node_id", nodeID))
	}
	if operation := lnccontext.GetOperation(ctx); operation != "" {
		fields = append(fields, zap.String("operation", operation))
	}
	if duration := lnccontext.GetDuration(ctx); duration > 0 {
		fields = append(fields, zap.Duration("duration", duration))
	}

	return fields
}

// Debug logs a debug message with context.
func (cl *ContextLogger) Debug(ctx context.Context, msg string,
	fields ...zap.Field) {
	cl.WithContext(ctx).Debug(msg, fields...)
}

// Info logs an info message with context.
func (cl *ContextLogger) Info(ctx context.Context, msg string,
	fields ...zap.Field) {
	cl.WithContext(ctx).Info(msg, fields...)
}

// Warn logs a warning message with context.
func (cl *ContextLogger) Warn(ctx context.Context, msg string,
	fields ...zap.Field) {
	cl.WithContext(ctx).Warn(msg, fields...)
}

// Error logs an error message with context.
func (cl *ContextLogger) Error(ctx context.Context, msg string,
	fields ...zap.Field) {
	cl.WithContext(ctx).Error(msg, fields...)
}

// Fatal logs a fatal message with context and exits.
func (cl *ContextLogger) Fatal(ctx context.Context, msg string,
	fields ...zap.Field) {
	cl.WithContext(ctx).Fatal(msg, fields...)
}

// DPanic logs a development panic message with context.
func (cl *ContextLogger) DPanic(ctx context.Context, msg string,
	fields ...zap.Field) {
	cl.WithContext(ctx).DPanic(msg, fields...)
}

// With creates a child logger with additional fields.
func (cl *ContextLogger) With(fields ...zap.Field) *ContextLogger {
	return &ContextLogger{
		logger: cl.logger.With(fields...),
	}
}

// Check returns a CheckedEntry if logging a message at the specified level.
// Is enabled.
func (cl *ContextLogger) Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry {
	return cl.logger.Check(lvl, msg)
}

// Sync flushes any buffered log entries.
func (cl *ContextLogger) Sync() error {
	return cl.logger.Sync()
}

// Global context logger instance.
var ContextLog *ContextLogger

// InitContextLogger initializes the global context logger.
func InitContextLogger() {
	if Logger == nil {
		// Initialize regular logger first if needed
		_ = InitLogger(true)
	}
	ContextLog = NewContextLogger(Logger)
}

// LogWithContext is a convenience function for logging with context.
func LogWithContext(ctx context.Context) *zap.Logger {
	if ContextLog == nil {
		InitContextLogger()
	}
	return ContextLog.WithContext(ctx)
}
