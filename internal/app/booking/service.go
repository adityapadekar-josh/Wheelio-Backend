package booking

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/email"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/user"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/app/vehicle"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/cryptokit"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/middleware"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/repository"
)

type service struct {
	bookingRepository repository.BookingRepository
	userService       user.Service
	vehicleService    vehicle.Service
	emailService      email.Service
}

type Service interface {
	CreateBooking(ctx context.Context, bookingData CreateBookingRequestBody) (newBooking Booking, err error)
	CancelBooking(ctx context.Context, bookingId int) (err error)
}

func NewService(bookingRepository repository.BookingRepository, userService user.Service, vehicleService vehicle.Service, emailService email.Service) Service {
	return &service{
		bookingRepository: bookingRepository,
		userService:       userService,
		vehicleService:    vehicleService,
		emailService:      emailService,
	}
}

func (s *service) CreateBooking(ctx context.Context, bookingData CreateBookingRequestBody) (newBooking Booking, err error) {
	userId, ok := ctx.Value(middleware.RequestContextUserIdKey).(int)

	if !ok {
		slog.Error("failed to retrieve user id from context")
		return Booking{}, apperrors.ErrInternalServer
	}

	err = bookingData.validate()
	if err != nil {
		slog.Error("booking details validation failed", "error", err)
		return Booking{}, apperrors.ErrInvalidRequestBody
	}

	vehicle, err := s.vehicleService.GetVehicleById(ctx, bookingData.VehicleId)
	if err != nil {
		slog.Error("failed to retrieve vehicle details", "error", err)
		return Booking{}, err
	}

	if vehicle.IsDeleted {
		slog.Error("vehicle is deleted thus cannot create booking")
		return Booking{}, apperrors.ErrVehicleNotFound
	}

	err = s.bookingRepository.VehicleBookingConflictCheck(ctx, nil, vehicle.Id, bookingData.ScheduledPickupTime, bookingData.ScheduledDropoffTime)
	if err != nil {
		slog.Error("failed to check booking slot availability", "error", err)
		return Booking{}, err
	}

	user, err := s.userService.GetUserById(ctx, userId)
	if err != nil {
		slog.Error("failed to get the user to create booking", "error", err)
		return Booking{}, err
	}

	tx, err := s.bookingRepository.BeginTx(ctx)
	if err != nil {
		slog.Error("failed to start booking creation", "error", err)
		return Booking{}, err
	}

	defer func() {
		if txErr := s.bookingRepository.HandleTransaction(ctx, tx, err); txErr != nil {
			slog.Error("failed to handle transaction", "error", txErr)
			err = txErr
		}
	}()

	bookingData.HostId = vehicle.HostId
	bookingData.SeekerId = user.Id
	bookingData.Status = Scheduled
	bookingData.ScheduledPickupTime = time.Date(
		bookingData.ScheduledPickupTime.Year(),
		bookingData.ScheduledPickupTime.Month(),
		bookingData.ScheduledPickupTime.Day(),
		0,
		0,
		0,
		0,
		bookingData.ScheduledPickupTime.Location(),
	).UTC()
	bookingData.ScheduledDropoffTime = time.Date(
		bookingData.ScheduledDropoffTime.Year(),
		bookingData.ScheduledDropoffTime.Month(),
		bookingData.ScheduledDropoffTime.Day(),
		23,
		59,
		59,
		999999,
		bookingData.ScheduledDropoffTime.Location(),
	).UTC()
	duration := bookingData.ScheduledDropoffTime.Sub(bookingData.ScheduledPickupTime)
	numberOfHours := math.Ceil(duration.Hours())
	bookingData.BookingAmount = numberOfHours * vehicle.RatePerHour
	bookingData.OverdueFeeRatePerHour = vehicle.OverdueFeeRatePerHour

	booking, err := s.bookingRepository.CreateBooking(ctx, tx, repository.CreateBookingRequestBody(bookingData))
	if err != nil {
		slog.Error("failed to create booking", "error", err)
		return Booking{}, err
	}

	otp, err := cryptokit.GenerateOTP()
	if err != nil {
		slog.Error("failed to generate secure otp", "error", err)
		return Booking{}, apperrors.ErrInternalServer
	}

	optTokenData := OtpToken{
		BookingId: booking.Id,
		Otp:       "",
		ExpiresAt: booking.ScheduledDropoffTime,
	}
	err = s.bookingRepository.CreateOtpToken(ctx, tx, repository.OtpToken(optTokenData))
	if err != nil {
		slog.Error("failed to create checkout otp token", "error", err)
		return Booking{}, err
	}

	emailBody := fmt.Sprintf(checkoutOtpEmailContent, user.Name, otp, booking.ScheduledDropoffTime.Local().String())

	err = s.emailService.SendEmail(user.Name, user.Email, "Vehicle Checkout OTP – Wheelio", emailBody)
	if err != nil {
		slog.Error("failed to checkout otp email", "error", err)
		return Booking{}, err
	}

	return Booking(booking), nil
}

func (s *service) CancelBooking(ctx context.Context, bookingId int) (err error) {
	err = s.bookingRepository.UpdateBookingStatus(ctx, nil, bookingId, Cancelled)
	if err != nil {
		slog.Error("failed to cancel the booking", "error", err)
		return err
	}

	return nil
}
