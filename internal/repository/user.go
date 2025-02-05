package repository

import (
	"context"
	"database/sql"
	"time"
)

type userRepository struct {
	DB *sql.DB
}

type UserRepository interface {
	GetUserById(ctx context.Context, userId int) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	CreateUser(ctx context.Context, userData CreateUserRequestBody, role string) (User, error)
	UpdateUserEmailVerifiedStatus(ctx context.Context, userId int) error
	CreateVerificationToken(ctx context.Context, userId int, token, tokenType string, expiresAt time.Time) (VerificationToken, error)
	GetVerificationTokenByToken(ctx context.Context, token string) (VerificationToken, error)
	DeleteVerificationTokenById(ctx context.Context, tokenId int) error
	UpdateUserPassword(ctx context.Context, userId int, password string) error
	UpdateUserRole(ctx context.Context, userId int, role string) error
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{DB: db}
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

func (ur *userRepository) CreateUser(ctx context.Context, userData CreateUserRequestBody, role string) (User, error) {
	var user User
	err := ur.DB.QueryRowContext(
		ctx,
		createUserQuery,
		userData.Name,
		userData.Email,
		userData.PhoneNumber,
		userData.Password,
		role,
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
		return User{}, err
	}

	return user, nil
}

func (ur *userRepository) GetUserById(ctx context.Context, userId int) (User, error) {
	var user User
	err := ur.DB.QueryRowContext(
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
		return User{}, err
	}

	return user, nil
}

func (ur *userRepository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := ur.DB.QueryRowContext(
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
		return User{}, err
	}

	return user, nil
}

func (ur *userRepository) UpdateUserEmailVerifiedStatus(ctx context.Context, userId int) error {
	_, err := ur.DB.ExecContext(ctx, updateUserEmailVerifiedStatusQuery, userId)
	if err != nil {
		return err
	}

	return nil
}

func (ur *userRepository) UpdateUserPassword(ctx context.Context, userId int, password string) error {
	_, err := ur.DB.ExecContext(ctx, updateUserPasswordQuery, password, userId)
	if err != nil {
		return err
	}

	return nil
}

func (ur *userRepository) UpdateUserRole(ctx context.Context, userId int, role string) error {
	_, err := ur.DB.ExecContext(ctx, updateUserRoleQuery, role, userId)
	if err != nil {
		return err
	}

	return nil
}

func (ur *userRepository) CreateVerificationToken(ctx context.Context, userId int, token, tokenType string, expiresAt time.Time) (VerificationToken, error) {
	var verificationToken VerificationToken
	err := ur.DB.QueryRowContext(
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
		return verificationToken, err
	}

	return verificationToken, nil
}

func (ur *userRepository) GetVerificationTokenByToken(ctx context.Context, token string) (VerificationToken, error) {
	var verificationToken VerificationToken
	err := ur.DB.QueryRowContext(
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
		return verificationToken, err
	}

	return verificationToken, nil
}

func (ur *userRepository) DeleteVerificationTokenById(ctx context.Context, tokenId int) error {
	_, err := ur.DB.ExecContext(ctx, deleteVerificationTokenByIdQuery, tokenId)
	if err != nil {
		return err
	}

	return nil
}
