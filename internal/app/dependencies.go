package app

import (
	"database/sql"

	"cloud.google.com/go/storage"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/booking"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/email"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/firebase"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/user"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/vehicle"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/repository"
)

type Dependencies struct {
	UserService    user.Service
	VehicleService vehicle.Service
	BookingService booking.Service
}

func InitDependencies(db *sql.DB, firebaseBucket *storage.BucketHandle) Dependencies {
	userRepository := repository.NewUserRepository(db)
	vehicleRepository := repository.NewVehicleRepository(db)
	bookingRepository := repository.NewBookingRepository(db)

	emailService := email.NewService()
	firebaseService := firebase.NewService(firebaseBucket)
	userService := user.NewService(userRepository, emailService)
	vehicleService := vehicle.NewService(vehicleRepository, firebaseService)
	bookingService := booking.NewService(bookingRepository, userService, vehicleService, emailService)

	return Dependencies{
		UserService:    userService,
		VehicleService: vehicleService,
		BookingService: bookingService,
	}
}
