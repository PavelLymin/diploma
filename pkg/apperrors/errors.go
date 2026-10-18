package apperrors

import "errors"

var (
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrNotFound       = errors.New("not_found")
	ErrInternalServer = errors.New("internal_server")
)
