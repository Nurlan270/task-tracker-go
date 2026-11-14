package errors

import "errors"

var (
	ErrInternalServer   = errors.New("http: internal server error")
	ErrBadRequest       = errors.New("http: bad request")
	ErrNotFound         = errors.New("http: resource not found")
	ErrMethodNotAllowed = errors.New("http: method not allowed")
	ErrUnauthorized     = errors.New("http: unauthorized")
	ErrForbidden        = errors.New("http: forbidden")
)
