// Copyright (c) 2025 Lightning Labs
// Distributed under the MIT license. See LICENSE for details.

// Package context provides request-scoped metadata and tracing helpers for the
// MCP LNC server.
package context

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ContextKey is a type for context keys to avoid collisions.
type contextKey string

const (
	// Context keys for request metadata.
	requestIDKey contextKey = "request_id"
	traceIDKey   contextKey = "trace_id"
	userIDKey    contextKey = "user_id"
	sessionIDKey contextKey = "session_id"
	nodeIDKey    contextKey = "node_id"
	operationKey contextKey = "operation"
	startTimeKey contextKey = "start_time"
	deadlineKey  contextKey = "deadline"
)

// RequestContext wraps a standard context with request-specific metadata.
type RequestContext struct {
	context.Context
	cancel    context.CancelFunc
	requestID string
	traceID   string
	userID    string
	sessionID string
	nodeID    string
	operation string
	startTime time.Time
	deadline  time.Time
}

// New creates a new RequestContext with generated identifiers and a timeout.
func New(parent context.Context, operation string, timeout time.Duration) *RequestContext {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	if timeout > 0 {
		ctx, cancel = context.WithTimeout(parent, timeout)
	} else {
		ctx, cancel = context.WithCancel(parent)
	}

	now := time.Now()
	rc := &RequestContext{
		Context:   ctx,
		cancel:    cancel,
		requestID: uuid.New().String(),
		traceID:   uuid.New().String(),
		operation: operation,
		startTime: now,
		deadline:  now.Add(timeout),
	}

	// Store values in underlying context for middleware compatibility.
	rc.Context = context.WithValue(rc.Context, requestIDKey, rc.requestID)
	rc.Context = context.WithValue(rc.Context, traceIDKey, rc.traceID)
	rc.Context = context.WithValue(rc.Context, operationKey, rc.operation)
	rc.Context = context.WithValue(rc.Context, startTimeKey, rc.startTime)
	rc.Context = context.WithValue(rc.Context, deadlineKey, rc.deadline)

	return rc
}

// WithTraceID clones a RequestContext while reusing an existing trace ID for
// distributed tracing.
func WithTraceID(parent context.Context, traceID, operation string,
	timeout time.Duration) *RequestContext {
	rc := New(parent, operation, timeout)
	rc.traceID = traceID
	rc.Context = context.WithValue(rc.Context, traceIDKey, traceID)
	return rc
}

// Cancel releases resources associated with the context.
func (rc *RequestContext) Cancel() {
	if rc == nil || rc.cancel == nil {
		return
	}
	rc.cancel()
	rc.cancel = nil
}

// Done returns a channel that is closed when work associated with the context completes.
func (rc *RequestContext) Done() <-chan struct{} {
	if rc == nil {
		return nil
	}
	return rc.Context.Done()
}

// Err returns the error associated with the context, if any.
func (rc *RequestContext) Err() error {
	if rc == nil {
		return nil
	}
	return rc.Context.Err()
}

// WithUser adds user information to the context.
func (rc *RequestContext) WithUser(userID, sessionID string) *RequestContext {
	rc.userID = userID
	rc.sessionID = sessionID
	rc.Context = context.WithValue(rc.Context, userIDKey, userID)
	rc.Context = context.WithValue(rc.Context, sessionIDKey, sessionID)
	return rc
}

// WithNode adds Lightning node information to the context.
func (rc *RequestContext) WithNode(nodeID string) *RequestContext {
	rc.nodeID = nodeID
	rc.Context = context.WithValue(rc.Context, nodeIDKey, nodeID)
	return rc
}

// RequestID returns the unique request identifier.
func (rc *RequestContext) RequestID() string {
	return rc.requestID
}

// TraceID returns the trace identifier for distributed tracing.
func (rc *RequestContext) TraceID() string {
	return rc.traceID
}

// UserID returns the user identifier if set.
func (rc *RequestContext) UserID() string {
	return rc.userID
}

// SessionID returns the session identifier if set.
func (rc *RequestContext) SessionID() string {
	return rc.sessionID
}

// NodeID returns the Lightning node identifier if set.
func (rc *RequestContext) NodeID() string {
	return rc.nodeID
}

// Operation returns the operation name.
func (rc *RequestContext) Operation() string {
	return rc.operation
}

// StartTime returns when the request started.
func (rc *RequestContext) StartTime() time.Time {
	return rc.startTime
}

// Duration returns how long the request has been running.
func (rc *RequestContext) Duration() time.Duration {
	return time.Since(rc.startTime)
}

// TimeRemaining returns the time remaining before deadline.
func (rc *RequestContext) TimeRemaining() time.Duration {
	return time.Until(rc.deadline)
}

// IsExpired checks if the context deadline has passed.
func (rc *RequestContext) IsExpired() bool {
	return time.Now().After(rc.deadline)
}

// Helper functions for extracting values from any context.

// GetRequestID extracts the request ID from any context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// GetTraceID extracts the trace ID from any context.
func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey).(string); ok {
		return id
	}
	return ""
}

// GetUserID extracts the user ID from any context.
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(userIDKey).(string); ok {
		return id
	}
	return ""
}

// GetSessionID extracts the session ID from any context.
func GetSessionID(ctx context.Context) string {
	if id, ok := ctx.Value(sessionIDKey).(string); ok {
		return id
	}
	return ""
}

// GetNodeID extracts the node ID from any context.
func GetNodeID(ctx context.Context) string {
	if id, ok := ctx.Value(nodeIDKey).(string); ok {
		return id
	}
	return ""
}

// GetOperation extracts the operation name from any context.
func GetOperation(ctx context.Context) string {
	if op, ok := ctx.Value(operationKey).(string); ok {
		return op
	}
	return ""
}

// GetStartTime extracts the start time from any context.
func GetStartTime(ctx context.Context) time.Time {
	if t, ok := ctx.Value(startTimeKey).(time.Time); ok {
		return t
	}
	return time.Time{}
}

// GetDuration calculates the duration from the stored start time in context.
func GetDuration(ctx context.Context) time.Duration {
	if t, ok := ctx.Value(startTimeKey).(time.Time); ok {
		return time.Since(t)
	}
	return 0
}

// Fields returns all context fields as a map for logging.
func (rc *RequestContext) Fields() map[string]any {
	fields := make(map[string]any)

	if rc.requestID != "" {
		fields["request_id"] = rc.requestID
	}
	if rc.traceID != "" {
		fields["trace_id"] = rc.traceID
	}
	if rc.userID != "" {
		fields["user_id"] = rc.userID
	}
	if rc.sessionID != "" {
		fields["session_id"] = rc.sessionID
	}
	if rc.nodeID != "" {
		fields["node_id"] = rc.nodeID
	}
	if rc.operation != "" {
		fields["operation"] = rc.operation
	}
	fields["duration_ms"] = rc.Duration().Milliseconds()
	fields["time_remaining_ms"] = rc.TimeRemaining().Milliseconds()

	return fields
}

// FromContext attempts to cast a context to RequestContext.
func FromContext(ctx context.Context) (*RequestContext, bool) {
	rc, ok := ctx.(*RequestContext)
	return rc, ok
}

// Ensure wraps a context as RequestContext if it is not already.
func Ensure(ctx context.Context, operation string) *RequestContext {
	if rc, ok := FromContext(ctx); ok {
		return rc
	}

	// Check if we have an existing trace ID to maintain.
	if traceID := GetTraceID(ctx); traceID != "" {
		return WithTraceID(ctx, traceID, operation, 30*time.Second)
	}

	// Create new context with default timeout.
	return New(ctx, operation, 30*time.Second)
}
