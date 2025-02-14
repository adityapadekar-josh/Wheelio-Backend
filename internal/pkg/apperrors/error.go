package apperrors

import (
	"errors"
	"net/http"
)

var (
	ErrInternalServer     = errors.New("an unexpected error occurred. please try again later")
	ErrInvalidRequestBody = errors.New("invalid or missing parameters in the request body")
	ErrFailedMarshal      = errors.New("failed to parse request body")

	ErrUnauthorizedAccess = errors.New("unauthorized. please provide a valid access token")
	ErrAccessForbidden    = errors.New("access forbidden")
	ErrInvalidToken       = errors.New("invalid or expired token")

	ErrUserNotFound            = errors.New("user not found")
	ErrEmailAlreadyRegistered  = errors.New("email is already registered")
	ErrInvalidLoginCredentials = errors.New("invalid email or password")
	ErrUserNotVerified         = errors.New("user account is not verified. please check your email")

	ErrJWTCreationFailed   = errors.New("failed to create jwt token")
	ErrTokenCreationFailed = errors.New("failed to create verification token")

	ErrEmailSendFailed = errors.New("failed to send email")

	ErrInvalidImageToLink = errors.New("no image found to link")
	ErrVehicleNotFound    = errors.New("vehicle not found")
)

func MapError(err error) (statusCode int, errMessage string) {
	switch err {
	case ErrInvalidRequestBody:
		return http.StatusBadRequest, err.Error()
	case ErrInvalidLoginCredentials, ErrUnauthorizedAccess:
		return http.StatusUnauthorized, err.Error()
	case ErrAccessForbidden:
		return http.StatusForbidden, err.Error()
	case ErrUserNotFound, ErrVehicleNotFound:
		return http.StatusNotFound, err.Error()
	case ErrEmailAlreadyRegistered, ErrUserNotVerified:
		return http.StatusConflict, err.Error()
	case ErrInvalidToken:
		return http.StatusUnprocessableEntity, err.Error()
	default:
		return http.StatusInternalServerError, ErrInternalServer.Error()
	}
}
