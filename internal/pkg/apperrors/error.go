package apperrors

import (
	"errors"
	"net/http"
)

var (
	ErrInternalServer     = errors.New("an unexpected error occurred. please try again later")
	ErrInvalidRequestBody = errors.New("invalid or missing parameters in the request body")
	ErrInvalidQueryParams = errors.New("invalid or missing query parameters")
	ErrFailedMarshal      = errors.New("failed to parse request body")

	ErrUnauthorizedAccess = errors.New("unauthorized. please provide a valid access token")
	ErrActionForbidden    = errors.New("action forbidden")
	ErrAccessForbidden    = errors.New("access forbidden")
	ErrInvalidToken       = errors.New("invalid or expired token")

	ErrUserNotFound            = errors.New("user not found")
	ErrEmailAlreadyRegistered  = errors.New("email is already registered")
	ErrInvalidLoginCredentials = errors.New("invalid email or password")
	ErrUserNotVerified         = errors.New("user account is not verified. please check your email")

	ErrJWTCreationFailed   = errors.New("failed to create jwt token")
	ErrTokenCreationFailed = errors.New("failed to create verification token")

	ErrEmailSendFailed = errors.New("failed to send email")

	ErrInvalidImageToLink   = errors.New("no image found to link")
	ErrVehicleNotFound      = errors.New("vehicle not found")
	ErrInvalidPickupDropoff = errors.New("pickup timestamp cannot be after dropoff timestamp")
	ErrInvalidPagination    = errors.New("page and limit must be greater than zero")

	ErrBookingConflict               = errors.New("booking slot is not available for the selected time range")
	ErrInvalidOtp                    = errors.New("invalid opt")
	ErrOptTokenNotFound              = errors.New("opt token not found")
	ErrBookingNotFound               = errors.New("booking not found")
	ErrBookingCancelled              = errors.New("cannot perform operations on cancelled booking")
	ErrBookingCancellationNotAllowed = errors.New("cancellation is not allowed for this booking")
)

func MapError(err error) (statusCode int, errMessage string) {
	switch err {
	case ErrInvalidRequestBody, ErrInvalidQueryParams, ErrInvalidPickupDropoff, ErrInvalidPagination, ErrOptTokenNotFound, ErrBookingNotFound:
		return http.StatusBadRequest, err.Error()
	case ErrUnauthorizedAccess:
		return http.StatusUnauthorized, err.Error()
	case ErrAccessForbidden, ErrActionForbidden, ErrBookingCancellationNotAllowed:
		return http.StatusForbidden, err.Error()
	case ErrUserNotFound, ErrVehicleNotFound:
		return http.StatusNotFound, err.Error()
	case ErrEmailAlreadyRegistered, ErrUserNotVerified, ErrBookingConflict, ErrInvalidOtp, ErrBookingCancelled:
		return http.StatusConflict, err.Error()
	case ErrInvalidToken, ErrInvalidLoginCredentials:
		return http.StatusUnprocessableEntity, err.Error()
	default:
		return http.StatusInternalServerError, ErrInternalServer.Error()
	}
}
