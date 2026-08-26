package domain

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrConflict       = errors.New("conflict")
	ErrInvalid        = errors.New("invalid request")
	ErrLeaseLost      = errors.New("lease lost")
	ErrNotImplemented = errors.New("not implemented")
)
