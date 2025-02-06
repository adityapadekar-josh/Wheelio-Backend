package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/email"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/config"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/cryptokit"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/middleware"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/repository"
	"github.com/golang-jwt/jwt/v5"
)

type service struct {
	userRepository repository.UserRepository
	emailService   email.Service
}

type Service interface {
	RegisterUser(ctx context.Context, userDetails CreateUserRequestBody) error
	LoginUser(ctx context.Context, loginDetails LoginUserRequestBody) (AccessToken, error)
	VerifyEmail(ctx context.Context, token Token) error
	ForgotPassword(ctx context.Context, email Email) error
	ResetPassword(ctx context.Context, resetPasswordDetails ResetPasswordRequestBody) error
	GetLoggedInUser(ctx context.Context) (User, error)
	UpgradeUserRoleToHost(ctx context.Context) error
}

func NewService(userRepository repository.UserRepository, emailService email.Service) Service {
	return &service{
		userRepository: userRepository,
		emailService:   emailService,
	}
}

func (s *service) RegisterUser(ctx context.Context, userDetails CreateUserRequestBody) error {
	err := userDetails.validate()
	if err != nil {
		slog.Error("user details validation failed", "error", err.Error())
		return apperrors.ErrInvalidRequestBody
	}

	_, err = s.userRepository.GetUserByEmail(ctx, userDetails.Email)
	if err == nil {
		slog.Error("attempted to register with an email that is already in use")
		return apperrors.ErrEmailAlreadyRegistered
	}

	hashedPassword, err := cryptokit.HashPassword(userDetails.Password)
	if err != nil {
		slog.Error("password hashing failed", "error", err)
		return apperrors.ErrInternalServer
	}

	userDetails.Password = hashedPassword

	newUser, err := s.userRepository.CreateUser(ctx, repository.CreateUserRequestBody(userDetails))
	if err != nil {
		slog.Error("failed to create new user", "error", err)
		return apperrors.ErrUserCreationFailed
	}

	token, err := cryptokit.GenerateSecureToken(64)
	if err != nil {
		slog.Error("failed to generate secure verification token", "error", err)
		return apperrors.ErrInternalServer
	}

	expiresAt := time.Now().Add(verificationTokenTTL)
	_, err = s.userRepository.CreateVerificationToken(ctx, newUser.Id, token, EmailVerification, expiresAt)
	if err != nil {
		slog.Error("failed to create verification token", "error", err)
		return apperrors.ErrTokenCreationFailed
	}

	cfg := config.GetConfig()
	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", cfg.ClientURL, token)
	emailBody := fmt.Sprintf(emailVerificationEmailContent, newUser.Name, verificationLink)

	err = s.emailService.SendEmail(newUser.Name, newUser.Email, "Action Required: Verify Your Wheelio Account", emailBody)
	if err != nil {
		slog.Error("failed to send verification email", "error", err)
		return apperrors.ErrEmailSendFailed
	}

	return nil
}

func (s *service) LoginUser(ctx context.Context, loginDetails LoginUserRequestBody) (AccessToken, error) {
	err := loginDetails.validate()
	if err != nil {
		slog.Error("failed to validate login details", "error", err)
		return AccessToken{}, apperrors.ErrInvalidRequestBody
	}

	user, err := s.userRepository.GetUserByEmail(ctx, loginDetails.Email)
	if err != nil {
		slog.Error("user not found by email", "error", err)
		return AccessToken{}, apperrors.ErrInvalidLoginCredentials
	}

	if !user.IsVerified {
		slog.Error("user is not verified", "email", loginDetails.Email)
		return AccessToken{}, apperrors.ErrUserNotVerified
	}

	isPasswordCorrect := cryptokit.CheckPasswordHash(loginDetails.Password, user.Password)
	if !isPasswordCorrect {
		slog.Error("invalid login password", "email", loginDetails.Email)
		return AccessToken{}, apperrors.ErrInvalidLoginCredentials
	}

	token, err := cryptokit.CreateJWTToken(jwt.MapClaims{
		"id":    user.Id,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(accessTokenTTL).Unix(),
	})
	if err != nil {
		slog.Error("failed to create jwt token", "error", err)
		return AccessToken{}, apperrors.ErrJWTCreationFailed
	}

	return AccessToken{AccessToken: token}, nil
}

func (s *service) VerifyEmail(ctx context.Context, token Token) error {
	err := token.validate()
	if err != nil {
		slog.Error("failed to validate token", "error", err)
		return apperrors.ErrInvalidRequestBody
	}

	verificationToken, err := s.userRepository.GetVerificationTokenByToken(ctx, token.Token)
	if err != nil || verificationToken.ExpiresAt.Before(time.Now()) || verificationToken.Type != EmailVerification {
		slog.Error("invalid or expired verification token", "error", err)
		return apperrors.ErrInvalidToken
	}

	err = s.userRepository.UpdateUserEmailVerifiedStatus(ctx, verificationToken.UserId)
	if err != nil {
		slog.Error("failed to update user email verified status", "userId", verificationToken.UserId, "error", err)
		return errors.New("failed to update user email verified status")
	}

	err = s.userRepository.DeleteVerificationTokenById(ctx, verificationToken.Id)
	if err != nil {
		slog.Warn("failed to delete verification token", "error", err)
	}

	return nil
}

func (s *service) ForgotPassword(ctx context.Context, email Email) error {
	err := email.validate()
	if err != nil {
		slog.Error("failed to validate email", "error", err)
		return apperrors.ErrInvalidRequestBody
	}

	user, err := s.userRepository.GetUserByEmail(ctx, email.Email)
	if err != nil {
		slog.Warn("no user found for the given email for password reset request", "email", email.Email)
		return nil
	}

	token, err := cryptokit.GenerateSecureToken(64)
	if err != nil {
		slog.Error("failed to generate secure token", "error", err)
		return apperrors.ErrInternalServer
	}

	expiresAt := time.Now().Add(verificationTokenTTL)
	_, err = s.userRepository.CreateVerificationToken(ctx, user.Id, token, PasswordReset, expiresAt)
	if err != nil {
		slog.Error("failed to create password reset token", "error", err)
		return apperrors.ErrTokenCreationFailed
	}

	cfg := config.GetConfig()
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", cfg.ClientURL, token)
	emailBody := fmt.Sprintf(resetPasswordEmailContent, user.Name, resetLink)

	err = s.emailService.SendEmail(user.Name, user.Email, "Reset Your Wheelio Password", emailBody)
	if err != nil {
		slog.Error("failed to send reset password email", "error", err)
		return apperrors.ErrEmailSendFailed
	}

	return nil
}

func (s *service) ResetPassword(ctx context.Context, resetPasswordDetails ResetPasswordRequestBody) error {
	err := resetPasswordDetails.validate()
	if err != nil {
		slog.Error("failed to validate reset password details", "error", err)
		return apperrors.ErrInvalidRequestBody
	}

	verificationToken, err := s.userRepository.GetVerificationTokenByToken(ctx, resetPasswordDetails.Token)
	if err != nil || verificationToken.ExpiresAt.Before(time.Now()) || verificationToken.Type != PasswordReset {
		slog.Error("invalid or expired reset password token", "error", err)
		return apperrors.ErrInvalidToken
	}

	_, err = s.userRepository.GetUserById(ctx, verificationToken.UserId)
	if err != nil {
		slog.Error("failed to find user by id", "error", err)
		return apperrors.ErrInvalidToken
	}

	hashedPassword, err := cryptokit.HashPassword(resetPasswordDetails.Password)
	if err != nil {
		slog.Error("failed to hash new password", "error", err)
		return apperrors.ErrInternalServer
	}

	err = s.userRepository.UpdateUserPassword(ctx, verificationToken.UserId, hashedPassword)
	if err != nil {
		slog.Error("failed to update user password", "error", err)
		return errors.New("failed to update user password")
	}

	err = s.userRepository.DeleteVerificationTokenById(ctx, verificationToken.Id)
	if err != nil {
		slog.Warn("failed to delete verification token", "error", err)
	}

	return nil
}

func (s *service) GetLoggedInUser(ctx context.Context) (User, error) {
	userId, ok := ctx.Value(middleware.RequestContextUserIdKey).(int)

	if !ok {
		slog.Error("failed to retrieve user id from context")
		return User{}, apperrors.ErrInternalServer
	}

	user, err := s.userRepository.GetUserById(ctx, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("user not found", "error", err)
			return User{}, apperrors.ErrUserNotFound
		}
		slog.Error("failed to get user by id", "error", err)
		return User{}, errors.New("failed to get logged in user details")
	}

	return User(user), nil
}

func (s *service) UpgradeUserRoleToHost(ctx context.Context) error {
	userId, ok := ctx.Value(middleware.RequestContextUserIdKey).(int)

	if !ok {
		slog.Error("failed to retrieve user id from context")
		return apperrors.ErrInternalServer
	}

	err := s.userRepository.UpdateUserRole(ctx, userId, Host)
	if err != nil {
		slog.Error("failed to upgrade user role to host", "error", err)
		return errors.New("failed to upgrade user role to host")
	}

	return nil
}
