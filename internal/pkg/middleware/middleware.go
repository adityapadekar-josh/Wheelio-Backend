package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/config"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/cryptokit"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/response"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

type RequestContextKey string

var RequestContextUserIdKey RequestContextKey = "userId"
var RequestContextRoleKey RequestContextKey = "role"

func ChainMiddleware(h http.HandlerFunc, m ...Middleware) http.HandlerFunc {
	wrapped := h

	for _, middleware := range m {
		wrapped = middleware(wrapped)
	}

	return wrapped
}

func AuthenticationMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bearerToken := r.Header.Get("Authorization")

		if bearerToken == "" {
			slog.Error("no authentication token provided in request")
			response.WriteJson(w, http.StatusUnauthorized, apperrors.ErrUnauthorizedAccess.Error(), nil)
			return
		}

		token := strings.Split(bearerToken, " ")[1]
		data, err := cryptokit.VerifyJWTToken(token)
		if err != nil {
			slog.Error("invalid or expired jwt token", "error", err)
			response.WriteJson(w, http.StatusUnauthorized, err.Error(), nil)
			return
		}

		userId, ok := data["id"].(float64)
		if !ok {
			slog.Error("user id missing or invalid in token", "token", token)
			response.WriteJson(w, http.StatusUnauthorized, apperrors.ErrUnauthorizedAccess.Error(), nil)
			return
		}

		role, ok := data["role"].(string)
		if !ok {
			slog.Error("role missing or invalid in token", "token", token)
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

			role, ok := ctx.Value(RequestContextRoleKey).(string)
			if !ok {
				slog.Error("role missing in context")
				response.WriteJson(w, http.StatusForbidden, apperrors.ErrAccessForbidden.Error(), nil)
				return
			}

			authorized := false

			for _, allowedRole := range allowedRoles {
				if role == allowedRole {
					authorized = true
					break
				}
			}

			if !authorized {
				slog.Error("unauthorized access attempt")
				response.WriteJson(w, http.StatusForbidden, apperrors.ErrAccessForbidden.Error(), nil)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := config.GetConfig()
		w.Header().Set("Access-Control-Allow-Origin", cfg.ClientURL)

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
