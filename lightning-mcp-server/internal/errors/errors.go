// Copyright (c) 2025 Lightning Labs
// Distributed under the MIT license. See LICENSE for details.

// Package errors defines structured error types used throughout the MCP LNC
// server. It mirrors LND patterns so callers gain more context than generic
// errors provide.
package errors

import (
	"fmt"
)

// ErrorCode represents different types of errors that can occur.
type ErrorCode uint32

const (
	// ErrCodeUnknown represents an unknown error.
	ErrCodeUnknown ErrorCode = 0

	// ErrCodeConnectionFailed represents a connection failure.
	ErrCodeConnectionFailed ErrorCode = 1

	// ErrCodeInvalidPairingPhrase represents an invalid pairing phrase.
	ErrCodeInvalidPairingPhrase ErrorCode = 2

	// ErrCodeTimeout represents a timeout error.
	ErrCodeTimeout ErrorCode = 3

	// ErrCodeNotConnected represents running operations without an active
	// connection.
	ErrCodeNotConnected ErrorCode = 4

	// ErrCodeInvalidInvoice represents an invalid invoice format.
	ErrCodeInvalidInvoice ErrorCode = 5

	// ErrCodeInsufficientBalance represents insufficient balance for the
	// requested operation.
	ErrCodeInsufficientBalance ErrorCode = 6

	// ErrCodeInvalidAddress represents an invalid address format.
	ErrCodeInvalidAddress ErrorCode = 7

	// ErrCodeServerShutdown represents server shutdown error.
	ErrCodeServerShutdown ErrorCode = 8
)

// String returns a human-readable description of the error code.
func (e ErrorCode) String() string {
	switch e {
	case ErrCodeUnknown:
		return "Unknown"
	case ErrCodeConnectionFailed:
		return "ConnectionFailed"
	case ErrCodeInvalidPairingPhrase:
		return "InvalidPairingPhrase"
	case ErrCodeTimeout:
		return "Timeout"
	case ErrCodeNotConnected:
		return "NotConnected"
	case ErrCodeInvalidInvoice:
		return "InvalidInvoice"
	case ErrCodeInsufficientBalance:
		return "InsufficientBalance"
	case ErrCodeInvalidAddress:
		return "InvalidAddress"
	case ErrCodeServerShutdown:
		return "ServerShutdown"
	default:
		return fmt.Sprintf("Unknown(%d)", uint32(e))
	}
}

// Error represents a structured error with code and context.
type Error struct {
	Code    ErrorCode
	Message string
	Cause   error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause error for error unwrapping.
func (e *Error) Unwrap() error {
	return e.Cause
}

// New creates a new structured error.
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Wrap creates a new structured error that wraps another error.
func Wrap(cause error, code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Wrapf creates a new structured error with formatted message that wraps
// another error.
func Wrapf(cause error, code ErrorCode, format string,
	args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Cause:   cause,
	}
}

// Is checks if the error has the given error code.
func Is(err error, code ErrorCode) bool {
	var e *Error
	if As(err, &e) {
		return e.Code == code
	}
	return false
}

// As attempts to cast the error to our Error type.
func As(err error, target **Error) bool {
	if err == nil {
		return false
	}

	if e, ok := err.(*Error); ok {
		*target = e
		return true
	}
	return false
}

// Common error constructors for frequently used errors.

// ErrConnectionFailed creates a connection failed error.
func ErrConnectionFailed(cause error, details string) *Error {
	return Wrap(cause, ErrCodeConnectionFailed,
		"failed to establish Lightning Network connection: "+details)
}

// ErrInvalidPairingPhrase creates an invalid pairing phrase error.
func ErrInvalidPairingPhrase(reason string) *Error {
	return New(ErrCodeInvalidPairingPhrase,
		"invalid pairing phrase: "+reason)
}

// ErrTimeout creates a timeout error.
func ErrTimeout(operation string) *Error {
	return New(ErrCodeTimeout,
		"operation timed out: "+operation)
}

// ErrNotConnected creates a not connected error.
func ErrNotConnected() *Error {
	return New(ErrCodeNotConnected,
		"not connected to Lightning node. Use lnc_connect first")
}

// ErrInvalidInvoice creates an invalid invoice error.
func ErrInvalidInvoice(reason string) *Error {
	return New(ErrCodeInvalidInvoice,
		"invalid invoice format: "+reason)
}

// ErrInsufficientBalance creates an insufficient balance error.
func ErrInsufficientBalance(required, available int64) *Error {
	return New(ErrCodeInsufficientBalance,
		fmt.Sprintf("insufficient balance: required %d, available %d",
			required, available))
}

// ErrInvalidAddress creates an invalid address error.
func ErrInvalidAddress(addr string) *Error {
	return New(ErrCodeInvalidAddress,
		"invalid address format: "+addr)
}
