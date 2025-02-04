package app

import (
	"database/sql"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/email"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/user"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/repository"
)

type Dependencies struct {
	UserService user.Service
}

func NewServices(db *sql.DB) Dependencies {
	emailService := email.NewService()

	userRepository := repository.NewUserRepository(db)

	userService := user.NewService(userRepository, emailService)

	return Dependencies{
		UserService: userService,
	}
}
