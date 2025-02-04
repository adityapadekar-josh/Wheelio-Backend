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

func (ur *userRepository) CreateUser(ctx context.Context, userData CreateUserRequestBody, role string) (User, error) {
	sqlStatement := `
	INSERT INTO users (name, email, phone_number, password, role)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING *`

	var user User
	err := ur.DB.QueryRowContext(
		ctx,
		sqlStatement,
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
	sqlStatement := "SELECT * FROM users WHERE id=$1"

	var user User
	err := ur.DB.QueryRowContext(
		ctx,
		sqlStatement,
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
	sqlStatement := "SELECT * FROM users WHERE email=$1"

	var user User
	err := ur.DB.QueryRowContext(
		ctx,
		sqlStatement,
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
	sqlStatement := "UPDATE users SET is_verified=true WHERE id=$1"

	_, err := ur.DB.ExecContext(ctx, sqlStatement, userId)
	if err != nil {
		return err
	}

	return nil
}

func (ur *userRepository) UpdateUserPassword(ctx context.Context, userId int, password string) error {
	sqlStatement := "UPDATE users SET password=$1 WHERE id=$2"

	_, err := ur.DB.ExecContext(ctx, sqlStatement, password, userId)
	if err != nil {
		return err
	}

	return nil
}

func (ur *userRepository) UpdateUserRole(ctx context.Context, userId int, role string) error {
	sqlStatement := "UPDATE users SET role=$1 WHERE id=$2"

	_, err := ur.DB.ExecContext(ctx, sqlStatement, role, userId)
	if err != nil {
		return err
	}

	return nil
}

func (ur *userRepository) CreateVerificationToken(ctx context.Context, userId int, token, tokenType string, expiresAt time.Time) (VerificationToken, error) {
	sqlStatement := `
	INSERT INTO verification_tokens (user_id, token, type, expires_at)
	VALUES ($1, $2, $3, $4)
	RETURNING *`

	var verificationToken VerificationToken
	err := ur.DB.QueryRowContext(
		ctx,
		sqlStatement,
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
	sqlStatement := "SELECT * FROM verification_tokens WHERE token=$1"

	var verificationToken VerificationToken
	err := ur.DB.QueryRowContext(
		ctx,
		sqlStatement,
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
	sqlStatement := "DELETE FROM verification_tokens WHERE id=$1"

	_, err := ur.DB.ExecContext(ctx, sqlStatement, tokenId)
	if err != nil {
		return err
	}

	return nil
}
