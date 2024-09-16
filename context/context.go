// Package context is a set of shorter names for the very stuttery context
// library.
package context

import (
	"context"
)

type (
	// T - context.Context
	T = context.Context
	// F - context.CancelFunc
	F = context.CancelFunc
	// C - context.CancelCauseFunc
	C = context.CancelCauseFunc
)

var (
	// Bg - context.Background
	Bg = context.Background
	// Cancel - context.WithCancel
	Cancel = context.WithCancel
	// Timeout - context.WithTimeout
	Timeout = context.WithTimeout
	// TimeoutCause - context.WithTimeoutCause
	TimeoutCause = context.WithTimeoutCause
	// TODO - context.TODO
	TODO = context.TODO
	// Value - context.WithValue
	Value = context.WithValue
	// CancelCause - context.WithCancelCause
	CancelCause = context.WithCancelCause
	// Canceled - context.Canceled
	Canceled = context.Canceled
)
