package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/user"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/constant"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/model"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/response"
	"github.com/go-playground/validator/v10"
)

func SignUpUser(userService user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var requestBody model.CreateUserRequestBody
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			slog.Error(constant.FAILED_MARSHAL, "error", err.Error())
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		validate := validator.New()
		err = validate.Struct(requestBody)
		if err != nil {
			slog.Error(constant.FAILED_VALIDATION, "error", err.Error())
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		err = userService.RegisterUser(ctx, requestBody, constant.SEEKER)

		if err != nil {
			slog.Error("failed to register new user", "error", err.Error())
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "user registered successfully. please check your email to verify your account.", nil)
	}
}

func SignInUser(userService user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var requestBody model.LoginUserRequestBody
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			slog.Error(constant.FAILED_MARSHAL, "error", err.Error())
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		validate := validator.New()
		err = validate.Struct(requestBody)
		if err != nil {
			slog.Error(constant.FAILED_VALIDATION, "error", err.Error())
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		loginData, err := userService.LoginUser(ctx, requestBody)
		if err != nil {
			slog.Error("failed to login user", "error", err.Error())
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "login successful", loginData)
	}
}

func VerifyEmail(userService user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var requestBody model.Token
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			slog.Error(constant.FAILED_MARSHAL, "error", err.Error())
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		err = userService.VerifyEmail(ctx, requestBody.Token)
		if err != nil {
			slog.Error("failed to verify user email", "error", err.Error())
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "email verification successful", nil)
	}
}

func ForgotPassword(userService user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var requestBody model.Email
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			slog.Error(constant.FAILED_MARSHAL, "error", err.Error())
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		validate := validator.New()
		err = validate.Struct(requestBody)
		if err != nil {
			slog.Error(constant.FAILED_VALIDATION, "error", err.Error())
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		err = userService.ForgotPassword(ctx, requestBody.Email)
		if err != nil {
			slog.Error("failed to generate forgot password link", "error", err.Error())
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "password reset instructions have been sent to your email", nil)
	}
}

func ResetPassword(userService user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var requestBody model.ResetPasswordRequestBody
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			slog.Error(constant.FAILED_MARSHAL, "error", err.Error())
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		validate := validator.New()
		err = validate.Struct(requestBody)
		if err != nil {
			slog.Error(constant.FAILED_VALIDATION, "error", err.Error())
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		err = userService.ResetPassword(ctx, requestBody)
		if err != nil {
			slog.Error("failed to reset user password", "error", err.Error())
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "password has been successfully reset", nil)
	}
}

func GetLoggedInUser(userService user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, err := userService.GetLoggedInUser(ctx)
		if err != nil {
			slog.Error("failed to get logged in user", "error", err.Error())
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "user details fetched successfully", user)
	}
}

func HostSignUpUser(userService user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var requestBody model.CreateUserRequestBody
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			slog.Error(constant.FAILED_MARSHAL, "error", err.Error())
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		validate := validator.New()
		err = validate.Struct(requestBody)
		if err != nil {
			slog.Error(constant.FAILED_VALIDATION, "error", err.Error())
			response.WriteJson(w, http.StatusBadRequest, apperrors.ErrInvalidRequestBody.Error(), nil)
			return
		}

		err = userService.RegisterUser(ctx, requestBody, constant.HOST)

		if err != nil {
			slog.Error("failed to register new host", "error", err.Error())
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "user registered successfully. please check your email to verify your account.", nil)
	}
}

func UpgradeUserRoleToHost(userService user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := userService.UpgradeUserRoleToHost(ctx)
		if err != nil {
			slog.Error("failed to upgrade user to host", "error", err.Error())
			status, errorMessage := apperrors.MapError(err)
			response.WriteJson(w, status, errorMessage, nil)
			return
		}

		response.WriteJson(w, http.StatusOK, "user successfully upgraded to HOST", nil)
	}
}
