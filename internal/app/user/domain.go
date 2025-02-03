package user

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/constant"
)

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
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	Password    string `json:"password"`
}

type LoginUserRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ResetPasswordRequestBody struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type AccessToken struct {
	AccessToken string `json:"accessToken"`
}

type Token struct {
	Token string `json:"token"`
}

type Email struct {
	Email string `json:"email"`
}

func (u *User) redactPassword() {
	u.Password = ""
}

func (c CreateUserRequestBody) validate() error {
	var validationErrors []string

	if strings.TrimSpace(c.Name) == "" {
		validationErrors = append(validationErrors, "name is required")
	}

	if strings.TrimSpace(c.Email) == "" {
		validationErrors = append(validationErrors, "email is required")
	} else if !regexp.MustCompile(constant.EmailRegex).MatchString(c.Email) {
		validationErrors = append(validationErrors, "invalid email format")
	}

	if strings.TrimSpace(c.PhoneNumber) == "" {
		validationErrors = append(validationErrors, "phone number is required")
	} else if !regexp.MustCompile(constant.PhoneRegex).MatchString(c.PhoneNumber) {
		validationErrors = append(validationErrors, "invalid phone number format")
	}

	if strings.TrimSpace(c.Password) == "" {
		validationErrors = append(validationErrors, "password is required")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

func (c LoginUserRequestBody) validate() error {
	var validationErrors []string

	if strings.TrimSpace(c.Email) == "" {
		validationErrors = append(validationErrors, "email is required")
	} else if !regexp.MustCompile(constant.EmailRegex).MatchString(c.Email) {
		validationErrors = append(validationErrors, "invalid email format")
	}

	if strings.TrimSpace(c.Password) == "" {
		validationErrors = append(validationErrors, "password is required")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

func (c ResetPasswordRequestBody) validate() error {
	var validationErrors []string

	if strings.TrimSpace(c.Token) == "" {
		validationErrors = append(validationErrors, "token is required")
	}

	if strings.TrimSpace(c.Password) == "" {
		validationErrors = append(validationErrors, "password is required")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

func (c Token) validate() error {
	var validationError string

	if strings.TrimSpace(c.Token) == "" {
		validationError = "token is required"
	}

	if validationError != "" {
		return fmt.Errorf("validation failed: %s", validationError)
	}

	return nil
}

func (c Email) validate() error {
	var validationError string

	if strings.TrimSpace(c.Email) == "" {
		validationError = "email is required"
	} else if !regexp.MustCompile(constant.EmailRegex).MatchString(c.Email) {
		validationError = "invalid email format"
	}

	if validationError != "" {
		return fmt.Errorf("validation failed: %s", validationError)
	}

	return nil
}
