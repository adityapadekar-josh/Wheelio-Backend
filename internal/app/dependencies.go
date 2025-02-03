package app

import (
	"database/sql"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/email"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/user"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/config"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/repository"
)

type Dependencies struct {
	UserService user.Service
}

func NewServices(db *sql.DB, cfg config.Config) Dependencies {
	emailService := email.NewService(cfg.EmailService.ApiKey, cfg.EmailService.FromName, cfg.EmailService.FromEmail)

	userRepository := repository.NewUserRepository(db)

	userService := user.NewService(userRepository, emailService)

	return Dependencies{
		UserService: userService,
	}
}
