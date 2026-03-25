package service

import "errors"

// Sentinel errors returned by all service methods.
// Controllers map these to HTTP status codes via errors.Is().
var (
	ErrNotFound = errors.New("not found")
	ErrInvalid  = errors.New("invalid input")
	ErrInternal = errors.New("internal error")
)
