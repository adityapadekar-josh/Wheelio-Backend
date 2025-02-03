package user

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/constant"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/cryptokit"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/email"
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
	VerifyEmail(ctx context.Context, token string) error
	ForgotPassword(ctx context.Context, email string) error
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

func (s *service) RegisterUser(ctx context.Context, userDetails CreateUserRequestBody, role string) error {
	_, err := s.userRepository.GetUserByEmail(ctx, userDetails.Email)

	if err == nil {
		return apperrors.CustomHTTPErr{
			StatusCode: http.StatusBadRequest,
			Message:    fmt.Sprintf("email is already registered: %s", userDetails.Email),
		}
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
	_, err = s.userRepository.CreateVerificationToken(ctx, newUser.Id, token, constant.EMAIL_VERIFICATION, expiresAt)
	if err != nil {
		return err
	}

	verificationLink := fmt.Sprintf("http://localhost:3000/verify-email?token=%s", token)
	emailBody := fmt.Sprintf(
		"Hello %s,\n\nThank you for registering on Wheelio. Please verify your email address by clicking the link below:\n\n%s\n\nThis link will expire in 10 minutes.\n\nBest regards,\nThe Wheelio Team",
		newUser.Name, verificationLink,
	)

	err = s.emailService.SendEmail(newUser.Name, newUser.Email, "Action Required: Verify Your Wheelio Account", emailBody)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) LoginUser(ctx context.Context, loginDetails LoginUserRequestBody) (AccessToken, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, loginDetails.Email)
	if err != nil {
		return AccessToken{}, apperrors.CustomHTTPErr{
			StatusCode: http.StatusUnauthorized,
			Message:    "invalid email or password",
		}
	}

	if !user.IsVerified {
		return AccessToken{}, apperrors.CustomHTTPErr{
			StatusCode: http.StatusConflict,
			Message:    "user account is not verified. please check your email",
		}
	}

	isPasswordCorrect := cryptokit.CheckPasswordHash(loginDetails.Password, user.Password)
	if !isPasswordCorrect {
		return AccessToken{}, apperrors.CustomHTTPErr{
			StatusCode: http.StatusUnauthorized,
			Message:    "invalid email or password",
		}
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

func (s *service) VerifyEmail(ctx context.Context, token string) error {
	verificationToken, err := s.userRepository.GetVerificationTokenByToken(ctx, token)
	if err != nil || verificationToken.ExpiresAt.Before(time.Now()) || verificationToken.Type != constant.EMAIL_VERIFICATION {
		return apperrors.CustomHTTPErr{
			StatusCode: http.StatusUnprocessableEntity,
			Message:    apperrors.ErrInvalidToken.Error(),
		}
	}

	err = s.userRepository.UpdateUserEmailVerifiedStatus(ctx, verificationToken.UserId)
	if err != nil {
		return err
	}

	s.userRepository.DeleteVerificationTokenById(ctx, verificationToken.Id)

	return nil
}

func (s *service) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepository.GetUserByEmail(ctx, email)

	if err != nil {
		return nil
	}

	token, err := cryptokit.GenerateSecureToken(64)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(10 * time.Minute)
	_, err = s.userRepository.CreateVerificationToken(ctx, user.Id, token, constant.PASSWORD_RESET, expiresAt)
	if err != nil {
		return err
	}

	resetLink := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", token)
	emailBody := fmt.Sprintf(
		"Hello %s,\n\nWe received a request to reset your password for your Wheelio account. Click the link below to set a new password:\n\n%s\n\nIf you did not request a password reset, please ignore this email. This link will expire in 10 minutes for security reasons.\n\nBest regards,\nThe Wheelio Team",
		user.Name, resetLink,
	)

	err = s.emailService.SendEmail(user.Name, user.Email, "Reset Your Wheelio Password", emailBody)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) ResetPassword(ctx context.Context, resetPasswordDetails ResetPasswordRequestBody) error {
	verificationToken, err := s.userRepository.GetVerificationTokenByToken(ctx, resetPasswordDetails.Token)

	if err != nil || verificationToken.ExpiresAt.Before(time.Now()) || verificationToken.Type != constant.PASSWORD_RESET {
		return apperrors.CustomHTTPErr{
			StatusCode: http.StatusUnprocessableEntity,
			Message:    apperrors.ErrInvalidToken.Error(),
		}
	}

	_, err = s.userRepository.GetUserById(ctx, verificationToken.UserId)
	if err != nil {
		return apperrors.CustomHTTPErr{
			StatusCode: http.StatusUnprocessableEntity,
			Message:    apperrors.ErrInvalidToken.Error(),
		}
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
	userId := ctx.Value("userId").(int)

	user, err := s.userRepository.GetUserById(ctx, userId)
	if err != nil {
		return User{}, apperrors.CustomHTTPErr{
			StatusCode: http.StatusUnprocessableEntity,
			Message:    apperrors.ErrUserNotFound.Error(),
		}
	}

	redactedUser := redactPassword(User(user))

	return redactedUser, nil
}

func (s *service) UpgradeUserRoleToHost(ctx context.Context) error {
	userId := ctx.Value("userId").(int)

	err := s.userRepository.UpdateUserRole(ctx, userId, constant.HOST)
	if err != nil {
		return err
	}

	return nil
}
