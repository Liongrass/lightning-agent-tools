// Copyright (c) 2025 Lightning Labs
// Distributed under the MIT license. See LICENSE for details.

package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test error code constants.
func TestErrorCodes(t *testing.T) {
	// Verify error codes are defined.
	assert.Equal(t, ErrorCode(0), ErrCodeUnknown)
	assert.Equal(t, ErrorCode(1), ErrCodeConnectionFailed)
	assert.Equal(t, ErrorCode(2), ErrCodeInvalidPairingPhrase)
	assert.Equal(t, ErrorCode(3), ErrCodeTimeout)
	assert.Equal(t, ErrorCode(4), ErrCodeNotConnected)
	assert.Equal(t, ErrorCode(5), ErrCodeInvalidInvoice)
	assert.Equal(t, ErrorCode(6), ErrCodeInsufficientBalance)
	assert.Equal(t, ErrorCode(7), ErrCodeInvalidAddress)
	assert.Equal(t, ErrorCode(8), ErrCodeServerShutdown)
}

// Test New function creates proper error.
func TestNew(t *testing.T) {
	err := New(ErrCodeTimeout, "test error message")

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "test error message")
	assert.Contains(t, err.Error(), "Timeout")

	// Test error structure.
	assert.Equal(t, ErrCodeTimeout, err.Code)
	assert.Equal(t, "test error message", err.Message)
	assert.Nil(t, err.Cause)
}

// Test Wrap function wraps existing errors.
func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrap(originalErr, ErrCodeConnectionFailed, "connection failed")

	assert.NotNil(t, wrappedErr)
	assert.Contains(t, wrappedErr.Error(), "connection failed")
	assert.Contains(t, wrappedErr.Error(), "ConnectionFailed")
	assert.Contains(t, wrappedErr.Error(), "original error")
	assert.Equal(t, originalErr, wrappedErr.Cause)
}

// Test Wrap with nil error.
func TestWrap_NilError(t *testing.T) {
	wrappedErr := Wrap(nil, ErrCodeUnknown, "test message")

	assert.NotNil(t, wrappedErr)
	assert.Contains(t, wrappedErr.Error(), "test message")
	assert.Contains(t, wrappedErr.Error(), "Unknown")
}

// Test error message formatting.
func TestErrorMessageFormatting(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		message  string
		expected string
	}{
		{
			name:     "basic_error",
			code:     ErrCodeInvalidInvoice,
			message:  "field is required",
			expected: "[InvalidInvoice] field is required",
		},
		{
			name:     "empty_message",
			code:     ErrCodeTimeout,
			message:  "",
			expected: "[Timeout] ",
		},
		{
			name:     "complex_message",
			code:     ErrCodeConnectionFailed,
			message:  "failed to connect to server at 127.0.0.1:8080",
			expected: "[ConnectionFailed] failed to connect to server at 127.0.0.1:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.code, tt.message)
			assert.Equal(t, tt.expected, err.Error())
		})
	}
}

// Test error wrapping chain.
func TestErrorWrappingChain(t *testing.T) {
	// Create a chain of wrapped errors.
	originalErr := errors.New("database connection failed")
	level1Err := Wrap(originalErr, ErrCodeConnectionFailed, "service unavailable")
	level2Err := Wrap(level1Err, ErrCodeUnknown, "request processing failed")

	// Verify all messages are present in the final error.
	finalError := level2Err.Error()
	assert.Contains(t, finalError, "database connection failed")
	assert.Contains(t, finalError, "service unavailable")
	assert.Contains(t, finalError, "request processing failed")
	assert.Contains(t, finalError, "ConnectionFailed")
	assert.Contains(t, finalError, "Unknown")
}

// Test Error implements error interface.
func TestErrorInterface(t *testing.T) {
	err := &Error{
		Code:    ErrCodeInvalidInvoice,
		Message: "test message",
		Cause:   nil,
	}

	// Should implement error interface.
	var _ error = err

	assert.Equal(t, "[InvalidInvoice] test message", err.Error())
}

// Test Error with cause.
func TestErrorWithCause(t *testing.T) {
	originalErr := errors.New("original")
	err := &Error{
		Code:    ErrCodeTimeout,
		Message: "timeout occurred",
		Cause:   originalErr,
	}

	errorStr := err.Error()
	assert.Contains(t, errorStr, "[Timeout] timeout occurred")
	assert.Contains(t, errorStr, "original")
}

// Test error codes are unique.
func TestErrorCodesUnique(t *testing.T) {
	codes := []ErrorCode{
		ErrCodeUnknown,
		ErrCodeConnectionFailed,
		ErrCodeInvalidPairingPhrase,
		ErrCodeTimeout,
		ErrCodeNotConnected,
		ErrCodeInvalidInvoice,
		ErrCodeInsufficientBalance,
		ErrCodeInvalidAddress,
		ErrCodeServerShutdown,
	}

	seen := make(map[ErrorCode]bool)
	for _, code := range codes {
		assert.False(t, seen[code], "Error code %v is not unique", code)
		seen[code] = true
	}
}

// Benchmark error creation.
func BenchmarkNew(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = New(ErrCodeInvalidInvoice, "benchmark error message")
	}
}

// Benchmark error wrapping.
func BenchmarkWrap(b *testing.B) {
	originalErr := errors.New("original error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Wrap(originalErr, ErrCodeConnectionFailed, "wrapped error")
	}
}

// Benchmark error string generation.
func BenchmarkErrorString(b *testing.B) {
	err := New(ErrCodeInvalidInvoice, "benchmark error message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

// Test Is function.
func TestIs(t *testing.T) {
	err := New(ErrCodeTimeout, "timeout error")

	assert.True(t, Is(err, ErrCodeTimeout))
	assert.False(t, Is(err, ErrCodeUnknown))

	// Test with non-Error type.
	stdErr := errors.New("standard error")
	assert.False(t, Is(stdErr, ErrCodeTimeout))

	// Test with nil.
	assert.False(t, Is(nil, ErrCodeTimeout))
}

// Test As function.
func TestAs(t *testing.T) {
	err := New(ErrCodeInvalidInvoice, "invalid invoice")

	var e *Error
	assert.True(t, As(err, &e))
	assert.Equal(t, ErrCodeInvalidInvoice, e.Code)
	assert.Equal(t, "invalid invoice", e.Message)

	// Test with non-Error type.
	stdErr := errors.New("standard error")
	var e2 *Error
	assert.False(t, As(stdErr, &e2))

	// Test with nil.
	var e3 *Error
	assert.False(t, As(nil, &e3))
}

// Test Unwrap method.
func TestUnwrap(t *testing.T) {
	originalErr := errors.New("original")
	wrappedErr := Wrap(originalErr, ErrCodeConnectionFailed, "wrapped")

	assert.Equal(t, originalErr, wrappedErr.Unwrap())

	// Test unwrapping with no cause
	errorWithoutCause := New(ErrCodeTimeout, "timeout")
	assert.Nil(t, errorWithoutCause.Unwrap())
}

// Test Wrapf function.
func TestWrapf(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrapf(originalErr, ErrCodeConnectionFailed, "connection failed to %s:%d", "localhost", 8080)

	assert.NotNil(t, wrappedErr)
	assert.Contains(t, wrappedErr.Error(), "connection failed to localhost:8080")
	assert.Contains(t, wrappedErr.Error(), "original error")
	assert.Equal(t, ErrCodeConnectionFailed, wrappedErr.Code)
	assert.Equal(t, originalErr, wrappedErr.Cause)
}

// Test error code String method.
func TestErrorCodeString(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected string
	}{
		{ErrCodeUnknown, "Unknown"},
		{ErrCodeConnectionFailed, "ConnectionFailed"},
		{ErrCodeInvalidPairingPhrase, "InvalidPairingPhrase"},
		{ErrCodeTimeout, "Timeout"},
		{ErrCodeNotConnected, "NotConnected"},
		{ErrCodeInvalidInvoice, "InvalidInvoice"},
		{ErrCodeInsufficientBalance, "InsufficientBalance"},
		{ErrCodeInvalidAddress, "InvalidAddress"},
		{ErrCodeServerShutdown, "ServerShutdown"},
		{ErrorCode(999), "Unknown(999)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.code.String())
		})
	}
}

// Test common error constructors.
func TestCommonErrorConstructors(t *testing.T) {
	t.Run("ErrConnectionFailed", func(t *testing.T) {
		originalErr := errors.New("connection refused")
		err := ErrConnectionFailed(originalErr, "server unreachable")

		assert.Equal(t, ErrCodeConnectionFailed, err.Code)
		assert.Contains(t, err.Message, "server unreachable")
		assert.Equal(t, originalErr, err.Cause)
	})

	t.Run("ErrInvalidPairingPhrase", func(t *testing.T) {
		err := ErrInvalidPairingPhrase("wrong number of words")

		assert.Equal(t, ErrCodeInvalidPairingPhrase, err.Code)
		assert.Contains(t, err.Message, "wrong number of words")
		assert.Nil(t, err.Cause)
	})

	t.Run("ErrTimeout", func(t *testing.T) {
		err := ErrTimeout("connection establishment")

		assert.Equal(t, ErrCodeTimeout, err.Code)
		assert.Contains(t, err.Message, "connection establishment")
		assert.Nil(t, err.Cause)
	})

	t.Run("ErrNotConnected", func(t *testing.T) {
		err := ErrNotConnected()

		assert.Equal(t, ErrCodeNotConnected, err.Code)
		assert.Contains(t, err.Message, "not connected")
		assert.Nil(t, err.Cause)
	})

	t.Run("ErrInvalidInvoice", func(t *testing.T) {
		err := ErrInvalidInvoice("invalid checksum")

		assert.Equal(t, ErrCodeInvalidInvoice, err.Code)
		assert.Contains(t, err.Message, "invalid checksum")
		assert.Nil(t, err.Cause)
	})

	t.Run("ErrInsufficientBalance", func(t *testing.T) {
		err := ErrInsufficientBalance(1000, 500)

		assert.Equal(t, ErrCodeInsufficientBalance, err.Code)
		assert.Contains(t, err.Message, "1000")
		assert.Contains(t, err.Message, "500")
		assert.Nil(t, err.Cause)
	})

	t.Run("ErrInvalidAddress", func(t *testing.T) {
		err := ErrInvalidAddress("invalid-address")

		assert.Equal(t, ErrCodeInvalidAddress, err.Code)
		assert.Contains(t, err.Message, "invalid-address")
		assert.Nil(t, err.Cause)
	})
}
