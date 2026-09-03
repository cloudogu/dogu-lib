package doguregistry

import "errors"

type errorType int

const (
	errNotFound errorType = iota + 1
	errConnection
	errGeneric
	errUnauthorized
	errForbidden
)

var _ error = Error{}

type Error struct {
	errType errorType
	cause   error
}

func (c Error) Error() string {
	if c.cause == nil {
		return ""
	}
	return c.cause.Error()
}

func (c Error) Unwrap() error {
	return c.cause
}

func NewGenericError(err error) Error {
	return Error{
		errType: errGeneric,
		cause:   err,
	}
}

func NewNotFoundError(err error) Error {
	return Error{
		errType: errNotFound,
		cause:   err,
	}
}

func NewConnectionError(err error) Error {
	return Error{
		errType: errConnection,
		cause:   err,
	}
}

func NewUnauthorizedError(err error) Error {
	return Error{
		errType: errUnauthorized,
		cause:   err,
	}
}

func NewForbiddenError(err error) Error {
	return Error{
		errType: errForbidden,
		cause:   err,
	}
}

func isError(err error, t errorType) bool {
	var e Error
	if ok := errors.As(err, &e); !ok {
		return false
	}

	if e.errType == t {
		return true
	}

	return false
}

func IsGenericError(err error) bool {
	return isError(err, errGeneric)
}

func IsNotFoundError(err error) bool {
	return isError(err, errNotFound)
}

func IsConnectionError(err error) bool {
	return isError(err, errConnection)
}

func IsUnauthorizedError(err error) bool {
	return isError(err, errUnauthorized)
}

func IsForbiddenError(err error) bool {
	return isError(err, errForbidden)
}
