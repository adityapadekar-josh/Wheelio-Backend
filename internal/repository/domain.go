package repository

import "time"

type User struct {
	Id          int
	Name        string
	Email       string
	PhoneNumber string
	Password    string
	Role        string
	IsVerified  bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateUserRequestBody struct {
	Name        string
	Email       string
	PhoneNumber string
	Password    string
	Role        string
}

type VerificationToken struct {
	Id        int
	UserId    int
	Token     string
	Type      string
	ExpiresAt time.Time
}
