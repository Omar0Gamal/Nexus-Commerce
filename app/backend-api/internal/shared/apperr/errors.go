// Package apperr defines reusable sentinel errors shared across domain packages.
package apperr

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrConflict      = errors.New("conflict")
	ErrInvalidUUID   = errors.New("invalid UUID")
	ErrForbidden     = errors.New("forbidden")
	ErrUnprocessable = errors.New("unprocessable")
)
