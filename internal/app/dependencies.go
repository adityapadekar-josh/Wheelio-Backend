package app

import (
	"database/sql"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/email"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/user"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/vehicle"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/repository"
)

type Dependencies struct {
	UserService    user.Service
	VehicleService vehicle.Service
}

func NewServices(db *sql.DB) Dependencies {
	userRepository := repository.NewUserRepository(db)
	vehicleRepository := repository.NewVehicleRepository(db)

	emailService := email.NewService()
	userService := user.NewService(userRepository, emailService)
	vehicleService := vehicle.NewService(vehicleRepository)

	return Dependencies{
		UserService:    userService,
		VehicleService: vehicleService,
	}
}
