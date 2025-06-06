// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package atomic

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorByValue(t *testing.T) {
	err := &Error{}
	require.Nil(t, err.Load(), "Initial value shall be nil")
}

func TestNewErrorWithNilArgument(t *testing.T) {
	err := NewError(nil)
	require.Nil(t, err.Load(), "Initial value shall be nil")
}

func TestErrorCanStoreNil(t *testing.T) {
	err := NewError(errors.New("hello"))
	err.Store(nil)
	require.Nil(t, err.Load(), "Stored value shall be nil")
}

func TestNewErrorWithError(t *testing.T) {
	err1 := errors.New("hello1")
	err2 := errors.New("hello2")

	atom := NewError(err1)
	require.Equal(t, err1, atom.Load(), "Expected Load to return initialized value")

	atom.Store(err2)
	require.Equal(t, err2, atom.Load(), "Expected Load to return overridden value")
}

func TestErrorSwap(t *testing.T) {
	err1 := errors.New("hello1")
	err2 := errors.New("hello2")

	atom := NewError(err1)
	require.Equal(t, err1, atom.Load(), "Expected Load to return initialized value")

	old := atom.Swap(err2)
	require.Equal(t, err2, atom.Load(), "Expected Load to return overridden value")
	require.Equal(t, err1, old, "Expected old to be initial value")
}

func TestErrorCompareAndSwap(t *testing.T) {
	err1 := errors.New("hello1")
	err2 := errors.New("hello2")

	atom := NewError(err1)
	require.Equal(t, err1, atom.Load(), "Expected Load to return initialized value")

	swapped := atom.CompareAndSwap(err2, err2)
	require.False(t, swapped, "Expected swapped to be false")
	require.Equal(t, err1, atom.Load(), "Expected Load to return initial value")

	swapped = atom.CompareAndSwap(err1, err2)
	require.True(t, swapped, "Expected swapped to be true")
	require.Equal(t, err2, atom.Load(), "Expected Load to return overridden value")
}

func TestError_InitializeDefaults(t *testing.T) {
	tests := []struct {
		msg      string
		newError func() *Error
	}{
		{
			msg: "Uninitialized",
			newError: func() *Error {
				var e Error
				return &e
			},
		},
		{
			msg: "NewError with default",
			newError: func() *Error {
				return NewError(nil)
			},
		},
		{
			msg: "Error swapped with default",
			newError: func() *Error {
				e := NewError(assert.AnError)
				_ = e.Swap(nil)
				return e
			},
		},
		{
			msg: "Error CAS'd with default",
			newError: func() *Error {
				e := NewError(assert.AnError)
				e.CompareAndSwap(assert.AnError, nil)
				return e
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			t.Run("CompareAndSwap", func(t *testing.T) {
				e := tt.newError()
				require.True(t, e.CompareAndSwap(nil, assert.AnError))
				assert.Equal(t, assert.AnError, e.Load())
			})

			t.Run("Swap", func(t *testing.T) {
				e := tt.newError()
				assert.Equal(t, nil, e.Swap(assert.AnError))
			})
		})
	}
}
