package apperrors

import (
	"errors"
)

var (
	ErrInvalidRequestBody = errors.New("validation failed. please check the provided details")
	ErrInternalServer     = errors.New("an unexpected error occurred. please try again later")
	ErrUnauthorizedAccess = errors.New("unauthorized. please provide a valid access token")
	ErrAccessForbidden    = errors.New("access forbidden")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrUserNotFound       = errors.New("user not found")
)

type CustomHTTPErr struct {
	StatusCode int
	Message    string
}

func (c CustomHTTPErr) Error() string {
	return c.Message
}

type RequestBodyValidationErr struct {
	Message string
}

func (r RequestBodyValidationErr) Error() string {
	return r.Message
}
