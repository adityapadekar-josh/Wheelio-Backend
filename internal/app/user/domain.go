package user

import "time"

type User struct {
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phoneNumber"`
	Password    string    `json:"password,omitempty"`
	Role        string    `json:"role"`
	IsVerified  bool      `json:"is_verified"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CreateUserRequestBody struct {
	Name        string `json:"name" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	PhoneNumber string `json:"phoneNumber" validate:"required"`
	Password    string `json:"password" validate:"required"`
}

type LoginUserRequestBody struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type ResetPasswordRequestBody struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AccessToken struct {
	AccessToken string `json:"accessToken"`
}

type Token struct {
	Token string `json:"token"`
}

type Email struct {
	Email string `json:"email" validate:"required,email"`
}

func (u *User) redactPassword() {
	u.Password = ""
}

func (c CreateUserRequestBody) validate() bool {
	if c.Name == "" {
		return false
	}

	if c.Email == "" {
		return false
	}

	if c.PhoneNumber == "" {
		return false
	}
	
	if c.Password == "" {
		return false
	}

	return true
}
