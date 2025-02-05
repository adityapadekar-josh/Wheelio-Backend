package apperrors

import (
	"errors"
	"net/http"
)

var (
	ErrInternalServer     = errors.New("an unexpected error occurred. please try again later")
	ErrInvalidRequestBody = errors.New("validation failed. please check the provided details")
	ErrFailedMarshal      = errors.New("failed to parse request body")

	ErrUnauthorizedAccess = errors.New("unauthorized. please provide a valid access token")
	ErrAccessForbidden    = errors.New("access forbidden")
	ErrInvalidToken       = errors.New("invalid or expired token")

	ErrUserNotFound            = errors.New("user not found")
	ErrEmailAlreadyRegistered  = errors.New("email is already registered")
	ErrInvalidLoginCredentials = errors.New("invalid email or password")
	ErrUserNotVerified         = errors.New("user account is not verified. please check your email")
)

type RequestBodyValidationErr struct {
	Message string
}

func (r RequestBodyValidationErr) Error() string {
	return r.Message
}

func MapError(err error) (statusCode int, errMessage string) {
	switch err.(type) {
	case RequestBodyValidationErr:
		return http.StatusBadRequest, ErrInvalidRequestBody.Error()
	}

	switch err {
	case ErrEmailAlreadyRegistered:
		return http.StatusBadRequest, err.Error()
	case ErrUserNotVerified:
		return http.StatusConflict, err.Error()
	case ErrInvalidLoginCredentials:
		return http.StatusUnauthorized, err.Error()
	case ErrInvalidToken:
		return http.StatusUnprocessableEntity, err.Error()
	default:
		return http.StatusInternalServerError, ErrInternalServer.Error()
	}
}
