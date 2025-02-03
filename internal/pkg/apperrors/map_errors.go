package apperrors

import (
	"net/http"
)

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
