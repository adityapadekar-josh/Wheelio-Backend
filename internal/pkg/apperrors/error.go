package apperrors

import (
	"errors"
	"net/http"
)

var (
	ErrInvalidRequestBody = errors.New("validation failed. please check the provided details")
	ErrInternalServer     = errors.New("an unexpected error occurred. please try again later")
	ErrUnauthorizedAccess = errors.New("unauthorized. please provide a valid access token")
	ErrAccessForbidden    = errors.New("access forbidden")
	ErrFailedMarshal      = errors.New("failed to parse request body")
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

func MapError(err error) (statusCode int, errMessage string) {
	switch e := err.(type) {
	case CustomHTTPErr:
		return e.StatusCode, e.Error()
	case RequestBodyValidationErr:
		return http.StatusBadRequest, ErrInvalidRequestBody.Error()
	default:
		return http.StatusInternalServerError, ErrInternalServer.Error()
	}
}
