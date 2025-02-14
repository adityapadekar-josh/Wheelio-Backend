package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
)

type userRepository struct {
	BaseRepository
}

type UserRepository interface {
	RepositoryTransaction
	GetUserById(ctx context.Context, tx *sql.Tx, userId int) (User, error)
	GetUserByEmail(ctx context.Context, tx *sql.Tx, email string) (User, error)
	CreateUser(ctx context.Context, tx *sql.Tx, userData CreateUserRequestBody) (User, error)
	UpdateUserEmailVerifiedStatus(ctx context.Context, tx *sql.Tx, userId int) error
	CreateVerificationToken(ctx context.Context, tx *sql.Tx, userId int, token, tokenType string, expiresAt time.Time) (VerificationToken, error)
	GetVerificationTokenByToken(ctx context.Context, tx *sql.Tx, token string) (VerificationToken, error)
	DeleteVerificationTokenById(ctx context.Context, tx *sql.Tx, tokenId int) error
	UpdateUserPassword(ctx context.Context, tx *sql.Tx, userId int, password string) error
	UpdateUserRole(ctx context.Context, tx *sql.Tx, userId int, role string) error
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{
		BaseRepository: BaseRepository{db},
	}
}

const (
	createUserQuery = `
	INSERT INTO users (name, email, phone_number, password, role)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING *`

	getUserByIdQuery = "SELECT * FROM users WHERE id=$1"

	getUserByEmailQuery = "SELECT * FROM users WHERE email=$1"

	updateUserEmailVerifiedStatusQuery = "UPDATE users SET is_verified=true WHERE id=$1"

	updateUserPasswordQuery = "UPDATE users SET password=$1 WHERE id=$2"

	updateUserRoleQuery = "UPDATE users SET role=$1 WHERE id=$2"

	createVerificationTokenQuery = `
	INSERT INTO verification_tokens (user_id, token, type, expires_at)
	VALUES ($1, $2, $3, $4)
	RETURNING *`

	getVerificationTokenByTokenQuery = "SELECT * FROM verification_tokens WHERE token=$1"

	deleteVerificationTokenByIdQuery = "DELETE FROM verification_tokens WHERE id=$1"
)

func (ur *userRepository) CreateUser(ctx context.Context, tx *sql.Tx, userData CreateUserRequestBody) (User, error) {
	executer := ur.initiateQueryExecuter(tx)

	var user User
	err := executer.QueryRowContext(
		ctx,
		createUserQuery,
		userData.Name,
		userData.Email,
		userData.PhoneNumber,
		userData.Password,
		userData.Role,
	).Scan(
		&user.Id,
		&user.Name,
		&user.Email,
		&user.PhoneNumber,
		&user.Password,
		&user.Role,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			slog.Error("attempted to register with an email that is already in use")
			return User{}, apperrors.ErrEmailAlreadyRegistered
		}
		slog.Error("failed to create user", "error", err)
		return User{}, apperrors.ErrInternalServer
	}

	return user, nil
}

func (ur *userRepository) GetUserById(ctx context.Context, tx *sql.Tx, userId int) (User, error) {
	executer := ur.initiateQueryExecuter(tx)

	var user User
	err := executer.QueryRowContext(
		ctx,
		getUserByIdQuery,
		userId,
	).Scan(
		&user.Id,
		&user.Name,
		&user.Email,
		&user.PhoneNumber,
		&user.Password,
		&user.Role,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("no user found with Id", "error", err)
			return User{}, apperrors.ErrUserNotFound
		}
		slog.Error("failed to fetch user with Id", "error", err)
		return User{}, apperrors.ErrInternalServer
	}

	return user, nil
}

func (ur *userRepository) GetUserByEmail(ctx context.Context, tx *sql.Tx, email string) (User, error) {
	executer := ur.initiateQueryExecuter(tx)

	var user User
	err := executer.QueryRowContext(
		ctx,
		getUserByEmailQuery,
		email,
	).Scan(
		&user.Id,
		&user.Name,
		&user.Email,
		&user.PhoneNumber,
		&user.Password,
		&user.Role,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("no user found with email", "error", err)
			return User{}, apperrors.ErrUserNotFound
		}
		slog.Error("failed to fetch user with email", "error", err)
		return User{}, apperrors.ErrInternalServer
	}

	return user, nil
}

func (ur *userRepository) UpdateUserEmailVerifiedStatus(ctx context.Context, tx *sql.Tx, userId int) error {
	executer := ur.initiateQueryExecuter(tx)

	_, err := executer.ExecContext(ctx, updateUserEmailVerifiedStatusQuery, userId)
	if err != nil {
		slog.Error("failed to update user verified status", "error", err)
		return apperrors.ErrInternalServer
	}

	return nil
}

func (ur *userRepository) UpdateUserPassword(ctx context.Context, tx *sql.Tx, userId int, password string) error {
	executer := ur.initiateQueryExecuter(tx)

	_, err := executer.ExecContext(ctx, updateUserPasswordQuery, password, userId)
	if err != nil {
		slog.Error("failed to update user password", "error", err)
		return apperrors.ErrInternalServer
	}

	return nil
}

func (ur *userRepository) UpdateUserRole(ctx context.Context, tx *sql.Tx, userId int, role string) error {
	executer := ur.initiateQueryExecuter(tx)

	_, err := executer.ExecContext(ctx, updateUserRoleQuery, role, userId)
	if err != nil {
		slog.Error("failed to update user role", "error", err)
		return apperrors.ErrInternalServer
	}

	return nil
}

func (ur *userRepository) CreateVerificationToken(ctx context.Context, tx *sql.Tx, userId int, token, tokenType string, expiresAt time.Time) (VerificationToken, error) {
	executer := ur.initiateQueryExecuter(tx)

	var verificationToken VerificationToken
	err := executer.QueryRowContext(
		ctx,
		createVerificationTokenQuery,
		userId,
		token,
		tokenType,
		expiresAt,
	).Scan(
		&verificationToken.Id,
		&verificationToken.UserId,
		&verificationToken.Token,
		&verificationToken.Type,
		&verificationToken.ExpiresAt,
	)
	if err != nil {
		slog.Error("failed to create verification token", "error", err)
		return VerificationToken{}, apperrors.ErrInternalServer
	}

	return verificationToken, nil
}

func (ur *userRepository) GetVerificationTokenByToken(ctx context.Context, tx *sql.Tx, token string) (VerificationToken, error) {
	executer := ur.initiateQueryExecuter(tx)

	var verificationToken VerificationToken
	err := executer.QueryRowContext(
		ctx,
		getVerificationTokenByTokenQuery,
		token,
	).Scan(
		&verificationToken.Id,
		&verificationToken.UserId,
		&verificationToken.Token,
		&verificationToken.Type,
		&verificationToken.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("no verification token found", "error", err)
			return VerificationToken{}, errors.New("no verification token found")
		}
		slog.Error("failed to fetch verification token", "error", err)
		return VerificationToken{}, apperrors.ErrInternalServer
	}

	return verificationToken, nil
}

func (ur *userRepository) DeleteVerificationTokenById(ctx context.Context, tx *sql.Tx, tokenId int) error {
	executer := ur.initiateQueryExecuter(tx)

	_, err := executer.ExecContext(ctx, deleteVerificationTokenByIdQuery, tokenId)
	if err != nil {
		slog.Error("failed to delete verification token", "error", err)
		return apperrors.ErrInternalServer
	}

	return nil
}
