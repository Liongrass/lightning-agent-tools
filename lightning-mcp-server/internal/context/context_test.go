// Copyright (c) 2025 Lightning Labs
// Distributed under the MIT license. See LICENSE for details.

package context

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test creating a new RequestContext.
func TestNew(t *testing.T) {
	ctx := New(context.Background(), "test_operation", 30*time.Second)
	defer ctx.Cancel()

	assert.NotNil(t, ctx)
	assert.NotEmpty(t, ctx.RequestID())
	assert.NotEmpty(t, ctx.TraceID())
	assert.Equal(t, "test_operation", ctx.Operation())
	assert.False(t, ctx.IsExpired())
	assert.True(t, ctx.TimeRemaining() > 0)
}

// Test creating RequestContext with existing trace ID.
func TestWithTraceID(t *testing.T) {
	existingTraceID := "trace-123"
	ctx := WithTraceID(context.Background(), existingTraceID,
		"test_operation", 30*time.Second)
	defer ctx.Cancel()

	assert.NotNil(t, ctx)
	assert.Equal(t, existingTraceID, ctx.TraceID())
	assert.NotEmpty(t, ctx.RequestID())
	assert.Equal(t, "test_operation", ctx.Operation())
}

// Test adding user information.
func TestWithUser(t *testing.T) {
	ctx := New(context.Background(), "test_operation", 30*time.Second)
	defer ctx.Cancel()
	ctx = ctx.WithUser("user-123", "session-456")

	assert.Equal(t, "user-123", ctx.UserID())
	assert.Equal(t, "session-456", ctx.SessionID())
}

// Test adding node information.
func TestWithNode(t *testing.T) {
	ctx := New(context.Background(), "test_operation", 30*time.Second)
	defer ctx.Cancel()
	ctx = ctx.WithNode("node-789")

	assert.Equal(t, "node-789", ctx.NodeID())
}

// Test duration calculation.
func TestDuration(t *testing.T) {
	ctx := New(context.Background(), "test_operation", 30*time.Second)
	defer ctx.Cancel()

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	duration := ctx.Duration()
	assert.True(t, duration >= 100*time.Millisecond)
	assert.True(t, duration < 200*time.Millisecond)
}

// Test timeout and expiration.
func TestTimeoutAndExpiration(t *testing.T) {
	// Create context with very short timeout
	ctx := New(context.Background(), "test_operation", 100*time.Millisecond)
	defer ctx.Cancel()

	assert.False(t, ctx.IsExpired())
	assert.True(t, ctx.TimeRemaining() > 0)

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	assert.True(t, ctx.IsExpired())
	assert.True(t, ctx.TimeRemaining() < 0)
}

// Test extracting values from context.
func TestExtractValues(t *testing.T) {
	ctx := New(context.Background(), "test_operation", 30*time.Second)
	defer ctx.Cancel()
	ctx = ctx.WithUser("user-123", "session-456")
	ctx = ctx.WithNode("node-789")

	// Extract values using static functions
	assert.NotEmpty(t, GetRequestID(ctx))
	assert.NotEmpty(t, GetTraceID(ctx))
	assert.Equal(t, "user-123", GetUserID(ctx))
	assert.Equal(t, "session-456", GetSessionID(ctx))
	assert.Equal(t, "node-789", GetNodeID(ctx))
	assert.Equal(t, "test_operation", GetOperation(ctx))
	assert.False(t, GetStartTime(ctx).IsZero())
	assert.True(t, GetDuration(ctx) >= 0)
}

// Test extracting from regular context returns empty.
func TestExtractFromRegularContext(t *testing.T) {
	ctx := context.Background()

	assert.Empty(t, GetRequestID(ctx))
	assert.Empty(t, GetTraceID(ctx))
	assert.Empty(t, GetUserID(ctx))
	assert.Empty(t, GetSessionID(ctx))
	assert.Empty(t, GetNodeID(ctx))
	assert.Empty(t, GetOperation(ctx))
	assert.True(t, GetStartTime(ctx).IsZero())
	assert.Equal(t, time.Duration(0), GetDuration(ctx))
}

// Test Fields method for logging.
func TestFields(t *testing.T) {
	ctx := New(context.Background(), "test_operation", 30*time.Second)
	defer ctx.Cancel()
	ctx = ctx.WithUser("user-123", "session-456")
	ctx = ctx.WithNode("node-789")

	fields := ctx.Fields()

	assert.NotEmpty(t, fields["request_id"])
	assert.NotEmpty(t, fields["trace_id"])
	assert.Equal(t, "user-123", fields["user_id"])
	assert.Equal(t, "session-456", fields["session_id"])
	assert.Equal(t, "node-789", fields["node_id"])
	assert.Equal(t, "test_operation", fields["operation"])
	assert.NotNil(t, fields["duration_ms"])
	assert.NotNil(t, fields["time_remaining_ms"])
}

// Test FromContext type assertion.
func TestFromContext(t *testing.T) {
	// Test with RequestContext
	reqCtx := New(context.Background(), "test_operation", 30*time.Second)
	defer reqCtx.Cancel()
	extracted, ok := FromContext(reqCtx)
	assert.True(t, ok)
	assert.Equal(t, reqCtx, extracted)

	// Test with regular context
	regularCtx := context.Background()
	extracted, ok = FromContext(regularCtx)
	assert.False(t, ok)
	assert.Nil(t, extracted)
}

// Test Ensure function.
func TestEnsure(t *testing.T) {
	t.Run("with_request_context", func(t *testing.T) {
		// Already a RequestContext
		reqCtx := New(context.Background(), "original", 30*time.Second)
		defer reqCtx.Cancel()
		ensured := Ensure(reqCtx, "new_operation")
		defer ensured.Cancel()

		// Should return the same context
		assert.Equal(t, reqCtx, ensured)
		assert.Equal(t, "original", ensured.Operation())
	})

	t.Run("with_regular_context", func(t *testing.T) {
		// Regular context
		regularCtx := context.Background()
		ensured := Ensure(regularCtx, "new_operation")
		defer ensured.Cancel()

		// Should create new RequestContext
		assert.NotNil(t, ensured)
		assert.Equal(t, "new_operation", ensured.Operation())
		assert.NotEmpty(t, ensured.RequestID())
		assert.NotEmpty(t, ensured.TraceID())
	})

	t.Run("with_existing_trace_id", func(t *testing.T) {
		// Regular context with trace ID
		regularCtx := context.Background()
		regularCtx = context.WithValue(regularCtx, traceIDKey, "existing-trace")
		ensured := Ensure(regularCtx, "new_operation")
		defer ensured.Cancel()

		// Should preserve trace ID
		assert.NotNil(t, ensured)
		assert.Equal(t, "existing-trace", ensured.TraceID())
		assert.Equal(t, "new_operation", ensured.Operation())
	})
}

// Test context cancellation propagation.
func TestContextCancellation(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	reqCtx := New(parent, "test_operation", 30*time.Second)
	defer reqCtx.Cancel()

	// Context should not be done initially
	select {
	case <-reqCtx.Done():
		t.Fatal("Context should not be done yet")
	default:
		// Expected
	}

	// Cancel parent
	cancel()

	// Context should be done now
	select {
	case <-reqCtx.Done():
		// Expected
		assert.Error(t, reqCtx.Err())
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context should be done after cancellation")
	}
}

// Test concurrent access.
func TestConcurrentAccess(t *testing.T) {
	ctx := New(context.Background(), "test_operation", 30*time.Second)
	ctx = ctx.WithUser("user-123", "session-456")

	done := make(chan bool, 10)

	// Multiple goroutines accessing context
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Read various fields
			_ = ctx.RequestID()
			_ = ctx.TraceID()
			_ = ctx.UserID()
			_ = ctx.Duration()
			_ = ctx.Fields()

			// Extract using static functions
			_ = GetRequestID(ctx)
			_ = GetTraceID(ctx)
			_ = GetUserID(ctx)
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for goroutine")
		}
	}
}

// Benchmark context creation.
func BenchmarkNew(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = New(context.Background(), "benchmark", 30*time.Second)
	}
}

// Benchmark field extraction.
func BenchmarkFieldExtraction(b *testing.B) {
	ctx := New(context.Background(), "benchmark", 30*time.Second)
	ctx = ctx.WithUser("user-123", "session-456")
	ctx = ctx.WithNode("node-789")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.Fields()
	}
}

// Benchmark static value extraction.
func BenchmarkStaticExtraction(b *testing.B) {
	ctx := New(context.Background(), "benchmark", 30*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetRequestID(ctx)
		_ = GetTraceID(ctx)
		_ = GetOperation(ctx)
	}
}
