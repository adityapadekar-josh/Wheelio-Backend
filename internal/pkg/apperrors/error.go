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
	ErrUserCreationFailed      = errors.New("failed to create user")

	ErrJWTCreationFailed   = errors.New("failed to create jwt token")
	ErrTokenCreationFailed = errors.New("failed to create verification token")

	ErrEmailSendFailed = errors.New("failed to send email")
)

func MapError(err error) (statusCode int, errMessage string) {
	switch err {
	case ErrEmailAlreadyRegistered, ErrInvalidRequestBody:
		return http.StatusBadRequest, err.Error()
	case ErrInvalidLoginCredentials, ErrUnauthorizedAccess:
		return http.StatusUnauthorized, err.Error()
	case ErrAccessForbidden:
		return http.StatusForbidden, err.Error()
	case ErrUserNotFound:
		return http.StatusNotFound, err.Error()
	case ErrUserNotVerified:
		return http.StatusConflict, err.Error()
	case ErrInvalidToken:
		return http.StatusUnprocessableEntity, err.Error()
	default:
		return http.StatusInternalServerError, ErrInternalServer.Error()
	}
}
