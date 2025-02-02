package apperrors

import (
	"net/http"
)

func MapError(err error) (statusCode int, errMessage string) {
	switch e := err.(type) {
	case CustomHTTPErr:
		return e.StatusCode, e.Error()
	default:
		return http.StatusInternalServerError, ErrInternalServer.Error()
	}
}
