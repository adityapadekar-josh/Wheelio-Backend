package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	RegisterUser(ctx context.Context, userDetails CreateUserRequestBody, role string) error
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

var cfg = config.GetConfig()

const (
	emailVerificationEmailContent = "Hello %s,\n\nThank you for registering on Wheelio. Please verify your email address by clicking the link below:\n\n%s\n\nThis link will expire in 10 minutes.\n\nBest regards,\nThe Wheelio Team"
	resetPasswordEmailContent     = "Hello %s,\n\nWe received a request to reset your password for your Wheelio account. Click the link below to set a new password:\n\n%s\n\nIf you did not request a password reset, please ignore this email. This link will expire in 10 minutes for security reasons.\n\nBest regards,\nThe Wheelio Team"
)

func (s *service) RegisterUser(ctx context.Context, userDetails CreateUserRequestBody, role string) error {
	err := userDetails.validate()
	if err != nil {
		return apperrors.RequestBodyValidationErr{
			Message: err.Error(),
		}
	}

	_, err = s.userRepository.GetUserByEmail(ctx, userDetails.Email)
	if err == nil {
		return apperrors.ErrEmailAlreadyRegistered
	}

	hashedPassword, err := cryptokit.HashPassword(userDetails.Password)
	if err != nil {
		return err
	}

	userDetails.Password = hashedPassword

	newUser, err := s.userRepository.CreateUser(ctx, repository.CreateUserRequestBody(userDetails), role)
	if err != nil {
		return err
	}

	token, err := cryptokit.GenerateSecureToken(64)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(10 * time.Minute)
	_, err = s.userRepository.CreateVerificationToken(ctx, newUser.Id, token, EmailVerification, expiresAt)
	if err != nil {
		return err
	}

	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", cfg.ClientURL, token)
	emailBody := fmt.Sprintf(emailVerificationEmailContent, newUser.Name, verificationLink)

	err = s.emailService.SendEmail(newUser.Name, newUser.Email, "Action Required: Verify Your Wheelio Account", emailBody)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) LoginUser(ctx context.Context, loginDetails LoginUserRequestBody) (AccessToken, error) {
	err := loginDetails.validate()
	if err != nil {
		return AccessToken{}, apperrors.RequestBodyValidationErr{
			Message: err.Error(),
		}
	}

	user, err := s.userRepository.GetUserByEmail(ctx, loginDetails.Email)
	if err != nil {
		return AccessToken{}, apperrors.ErrInvalidLoginCredentials
	}

	if !user.IsVerified {
		return AccessToken{}, apperrors.ErrUserNotVerified
	}

	isPasswordCorrect := cryptokit.CheckPasswordHash(loginDetails.Password, user.Password)
	if !isPasswordCorrect {
		return AccessToken{}, apperrors.ErrInvalidLoginCredentials
	}

	token, err := cryptokit.CreateJWTToken(jwt.MapClaims{
		"id":    user.Id,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(time.Hour * 24 * 30).Unix(),
	})
	if err != nil {
		return AccessToken{}, err
	}

	return AccessToken{AccessToken: token}, nil
}

func (s *service) VerifyEmail(ctx context.Context, token Token) error {
	err := token.validate()
	if err != nil {
		return apperrors.RequestBodyValidationErr{
			Message: err.Error(),
		}
	}

	verificationToken, err := s.userRepository.GetVerificationTokenByToken(ctx, token.Token)
	if err != nil || verificationToken.ExpiresAt.Before(time.Now()) || verificationToken.Type != EmailVerification {
		return apperrors.ErrInvalidToken
	}

	err = s.userRepository.UpdateUserEmailVerifiedStatus(ctx, verificationToken.UserId)
	if err != nil {
		return err
	}

	s.userRepository.DeleteVerificationTokenById(ctx, verificationToken.Id)

	return nil
}

func (s *service) ForgotPassword(ctx context.Context, email Email) error {
	err := email.validate()
	if err != nil {
		return apperrors.RequestBodyValidationErr{
			Message: err.Error(),
		}
	}

	user, err := s.userRepository.GetUserByEmail(ctx, email.Email)
	if err != nil {
		return nil
	}

	token, err := cryptokit.GenerateSecureToken(64)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(10 * time.Minute)
	_, err = s.userRepository.CreateVerificationToken(ctx, user.Id, token, PasswordReset, expiresAt)
	if err != nil {
		return err
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", cfg.ClientURL, token)
	emailBody := fmt.Sprintf(resetPasswordEmailContent, user.Name, resetLink)

	err = s.emailService.SendEmail(user.Name, user.Email, "Reset Your Wheelio Password", emailBody)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) ResetPassword(ctx context.Context, resetPasswordDetails ResetPasswordRequestBody) error {
	err := resetPasswordDetails.validate()
	if err != nil {
		return apperrors.RequestBodyValidationErr{
			Message: err.Error(),
		}
	}

	verificationToken, err := s.userRepository.GetVerificationTokenByToken(ctx, resetPasswordDetails.Token)
	if err != nil || verificationToken.ExpiresAt.Before(time.Now()) || verificationToken.Type != PasswordReset {
		return apperrors.ErrInvalidToken
	}

	_, err = s.userRepository.GetUserById(ctx, verificationToken.UserId)
	if err != nil {
		return apperrors.ErrInvalidToken
	}

	hashedPassword, err := cryptokit.HashPassword(resetPasswordDetails.Password)
	if err != nil {
		return err
	}

	err = s.userRepository.UpdateUserPassword(ctx, verificationToken.UserId, hashedPassword)
	if err != nil {
		return err
	}

	s.userRepository.DeleteVerificationTokenById(ctx, verificationToken.Id)

	return nil
}

func (s *service) GetLoggedInUser(ctx context.Context) (User, error) {
	userId, ok := ctx.Value(middleware.RequestContextUserIdKey).(int)

	if !ok {
		return User{}, apperrors.ErrInternalServer
	}

	user, err := s.userRepository.GetUserById(ctx, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, apperrors.ErrUserNotFound
		}
		return User{}, err
	}

	maskedUser := User(user)
	maskedUser.redactPassword()

	return maskedUser, nil
}

func (s *service) UpgradeUserRoleToHost(ctx context.Context) error {
	userId, ok := ctx.Value(middleware.RequestContextUserIdKey).(int)

	if !ok {
		return apperrors.ErrInternalServer
	}

	err := s.userRepository.UpdateUserRole(ctx, userId, Host)
	if err != nil {
		return err
	}

	return nil
}
