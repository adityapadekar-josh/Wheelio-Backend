package user

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/response"
)

func SignUpUser(userService Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var requestBody CreateUserRequestBody
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			slog.Error(apperrors.ErrFailedMarshal.Error(), "error", err)
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		err = userService.RegisterUser(ctx, requestBody)
		if err != nil {
			slog.Error("failed to register new user", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "user registered successfully. please check your email to verify your account.", nil)
	}
}

func SignInUser(userService Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var requestBody LoginUserRequestBody
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			slog.Error(apperrors.ErrFailedMarshal.Error(), "error", err)
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		loginData, err := userService.LoginUser(ctx, requestBody)
		if err != nil {
			slog.Error("failed to login user", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "login successful", loginData)
	}
}

func VerifyEmail(userService Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var requestBody Token
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			slog.Error(apperrors.ErrFailedMarshal.Error(), "error", err)
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		err = userService.VerifyEmail(ctx, requestBody)
		if err != nil {
			slog.Error("failed to verify user email", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "email verification successful", nil)
	}
}

func ForgotPassword(userService Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var requestBody Email
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			slog.Error(apperrors.ErrFailedMarshal.Error(), "error", err)
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		err = userService.ForgotPassword(ctx, requestBody)
		if err != nil {
			slog.Error("failed to generate forgot password link", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "password reset instructions have been sent to your email", nil)
	}
}

func ResetPassword(userService Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var requestBody ResetPasswordRequestBody
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			slog.Error(apperrors.ErrFailedMarshal.Error(), "error", err)
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		err = userService.ResetPassword(ctx, requestBody)
		if err != nil {
			slog.Error("failed to reset user password", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "password has been successfully reset", nil)
	}
}

func GetLoggedInUser(userService Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, err := userService.GetLoggedInUser(ctx)
		if err != nil {
			slog.Error("failed to get logged in user", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "user details fetched successfully", user)
	}
}

func UpgradeUserRoleToHost(userService Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := userService.UpgradeUserRoleToHost(ctx)
		if err != nil {
			slog.Error("failed to upgrade user to host", "error", err)
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "user successfully upgraded to HOST", nil)
	}
}
