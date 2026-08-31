package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError(assert.AnError)

	assert.Equal(t, assert.AnError, err.cause)
	assert.Equal(t, errNotFound, err.errType)
}

func TestNewConnectionError(t *testing.T) {
	err := NewConnectionError(assert.AnError)

	assert.Equal(t, assert.AnError, err.cause)
	assert.Equal(t, errConnection, err.errType)
}

func TestNewGenericError(t *testing.T) {
	err := NewGenericError(assert.AnError)

	assert.Equal(t, assert.AnError, err.cause)
	assert.Equal(t, errGeneric, err.errType)
}

func TestNewUnauthorizedErrorError(t *testing.T) {
	err := NewUnauthorizedError(assert.AnError)

	assert.Equal(t, assert.AnError, err.cause)
	assert.Equal(t, errUnauthorized, err.errType)
}

func TestNewForbiddenErrorWatchError(t *testing.T) {
	err := NewForbiddenError(assert.AnError)

	assert.Equal(t, assert.AnError, err.cause)
	assert.Equal(t, errForbidden, err.errType)
}

func TestError_Error(t *testing.T) {
	err := Error{
		errType: 0,
		cause:   assert.AnError,
	}

	assert.Equal(t, assert.AnError.Error(), err.Error())
}

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		xResult bool
	}{
		{
			name:    "NotFoundError",
			err:     NewNotFoundError(assert.AnError),
			xResult: true,
		},
		{
			name:    "No config error",
			err:     assert.AnError,
			xResult: false,
		},
		{
			name:    "error is nil",
			err:     nil,
			xResult: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.xResult, IsNotFoundError(tc.err))
		})
	}
}

func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		xResult bool
	}{
		{
			name:    "ConnectionError",
			err:     NewConnectionError(assert.AnError),
			xResult: true,
		},
		{
			name:    "NotFoundError",
			err:     NewNotFoundError(assert.AnError),
			xResult: false,
		},
		{
			name:    "No config error",
			err:     assert.AnError,
			xResult: false,
		},
		{
			name:    "error is nil",
			err:     nil,
			xResult: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.xResult, IsConnectionError(tc.err))
		})
	}
}

func TestIsGenericError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		xResult bool
	}{
		{
			name:    "GenericError",
			err:     NewGenericError(assert.AnError),
			xResult: true,
		},
		{
			name:    "NotFoundError",
			err:     NewNotFoundError(assert.AnError),
			xResult: false,
		},
		{
			name:    "No config error",
			err:     assert.AnError,
			xResult: false,
		},
		{
			name:    "error is nil",
			err:     nil,
			xResult: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.xResult, IsGenericError(tc.err))
		})
	}
}

func TestIsUnauthorizedError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		xResult bool
	}{
		{
			name:    "UnauthorizedError",
			err:     NewUnauthorizedError(assert.AnError),
			xResult: true,
		},
		{
			name:    "NotFoundError",
			err:     NewNotFoundError(assert.AnError),
			xResult: false,
		},
		{
			name:    "No config error",
			err:     assert.AnError,
			xResult: false,
		},
		{
			name:    "error is nil",
			err:     nil,
			xResult: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.xResult, IsUnauthorizedError(tc.err))
		})
	}
}

func TestIsForbiddenError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		xResult bool
	}{
		{
			name:    "ForbiddenError",
			err:     NewForbiddenError(assert.AnError),
			xResult: true,
		},
		{
			name:    "NotFoundError",
			err:     NewNotFoundError(assert.AnError),
			xResult: false,
		},
		{
			name:    "No config error",
			err:     assert.AnError,
			xResult: false,
		},
		{
			name:    "error is nil",
			err:     nil,
			xResult: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.xResult, IsForbiddenError(tc.err))
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	sut := Error{cause: assert.AnError}
	assert.ErrorIs(t, sut, assert.AnError)
}
