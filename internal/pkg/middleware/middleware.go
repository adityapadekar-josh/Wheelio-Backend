package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/cryptokit"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/response"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

func ChainMiddleware(h http.HandlerFunc, m ...Middleware) http.HandlerFunc {
	wrapped := h

	for _, middleware := range m {
		wrapped = middleware(wrapped)
	}

	return wrapped
}

type RequestContextKey string

var RequestContextUserIdKey RequestContextKey = "userId"
var RequestContextRoleKey RequestContextKey = "role"

func AuthenticationMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bearerToken := r.Header.Get("Authorization")

		if bearerToken == "" {
			response.WriteJson(w, http.StatusUnauthorized, apperrors.ErrUnauthorizedAccess.Error(), nil)
			return
		}

		token := strings.Split(bearerToken, " ")[1]
		data, err := cryptokit.VerifyJWTToken(token)
		if err != nil {
			response.WriteJson(w, http.StatusUnauthorized, err.Error(), nil)
			return
		}

		userId, ok := data["id"].(float64)
		if !ok {
			response.WriteJson(w, http.StatusUnauthorized, apperrors.ErrUnauthorizedAccess.Error(), nil)
			return
		}

		role, ok := data["role"].(string)
		if !ok {
			response.WriteJson(w, http.StatusUnauthorized, apperrors.ErrUnauthorizedAccess.Error(), nil)
			return
		}

		ctx := context.WithValue(r.Context(), RequestContextUserIdKey, int(userId))
		ctx = context.WithValue(ctx, RequestContextRoleKey, role)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	}
}

func AuthorizationMiddleware(allowedRoles ...string) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			role := ctx.Value(RequestContextRoleKey).(string)

			authorized := false

			for _, allowedRole := range allowedRoles {
				if role == allowedRole {
					authorized = true
					break
				}
			}

			if !authorized {
				response.WriteJson(w, http.StatusForbidden, apperrors.ErrAccessForbidden.Error(), nil)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}
