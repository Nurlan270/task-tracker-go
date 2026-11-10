package errors

import "errors"

var (
	InternalServer   = errors.New("http: internal server error")
	BadRequest       = errors.New("http: bad request")
	NotFound         = errors.New("http: resource not found")
	MethodNotAllowed = errors.New("http: method not allowed")
	Unauthorized     = errors.New("http: unauthorized")
	Forbidden        = errors.New("http: forbidden")
)
