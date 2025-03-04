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
	ConfirmPickup(ctx context.Context, bookingId int, otpData OtpRequestBody) (err error)
	InitiateReturn(ctx context.Context, bookingId int) (err error)
	ConfirmReturn(ctx context.Context, bookingId int, otpData OtpRequestBody) (err error)
	GetSeekerBookings(ctx context.Context, page, limit int) (bookings PaginatedBookingData, err error)
	GetHostBookings(ctx context.Context, page, limit int) (bookings PaginatedBookingData, err error)
	GetBookingDetailsById(ctx context.Context, bookingId int) (booking BookingDetails, err error)
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
	duration := bookingData.ScheduledDropoffTime.Sub(bookingData.ScheduledPickupTime)
	numberOfHours := math.Ceil(duration.Hours())
	bookingData.BookingAmount = numberOfHours * vehicle.RatePerHour
	bookingData.OverdueFeeRatePerHour = vehicle.OverdueFeeRatePerHour
	bookingData.CancellationAllowed = vehicle.CancellationAllowed

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
		Otp:       otp,
		ExpiresAt: booking.ScheduledDropoffTime,
	}
	err = s.bookingRepository.CreateOtpToken(ctx, tx, repository.OtpToken(optTokenData))
	if err != nil {
		slog.Error("failed to create checkout otp token", "error", err)
		return Booking{}, err
	}

	emailBody := fmt.Sprintf(checkoutOtpEmailContent, user.Name, otp)

	err = s.emailService.SendEmail(user.Name, user.Email, "Vehicle Checkout OTP – Wheelio", emailBody)
	if err != nil {
		slog.Error("failed to send checkout otp email", "error", err)
		return Booking{}, err
	}

	return Booking(booking), nil
}

func (s *service) CancelBooking(ctx context.Context, bookingId int) (err error) {
	userId, ok := ctx.Value(middleware.RequestContextUserIdKey).(int)
	if !ok {
		slog.Error("failed to retrieve user id from context")
		return apperrors.ErrInternalServer
	}

	booking, err := s.bookingRepository.GetBookingById(ctx, nil, bookingId)
	if err != nil {
		slog.Error("failed to get booking", "error", err)
		return err
	}

	if booking.HostId != userId && booking.SeekerId != userId {
		slog.Error("unauthorized booking cancellation attempt")
		return apperrors.ErrActionForbidden
	}

	if booking.SeekerId == userId && !booking.CancellationAllowed {
		slog.Error("cancellation not allowed for this booking")
		return apperrors.ErrBookingCancellationNotAllowed
	}

	err = s.bookingRepository.UpdateBookingStatus(ctx, nil, bookingId, Cancelled)
	if err != nil {
		slog.Error("failed to cancel the booking", "error", err)
		return err
	}

	return nil
}

func (s *service) ConfirmPickup(ctx context.Context, bookingId int, otpData OtpRequestBody) (err error) {
	userId, ok := ctx.Value(middleware.RequestContextUserIdKey).(int)
	if !ok {
		slog.Error("failed to retrieve user id from context")
		return apperrors.ErrInternalServer
	}

	booking, err := s.bookingRepository.GetBookingById(ctx, nil, bookingId)
	if err != nil {
		slog.Error("failed to get booking", "error", err)
		return err
	}

	if booking.HostId != userId || booking.Status != Scheduled {
		slog.Error("invalid booking pickup attempt")
		return apperrors.ErrActionForbidden
	}

	otpToken, err := s.bookingRepository.GetOtpToken(ctx, nil, otpData.Otp)
	if err != nil {
		slog.Error("failed to get otp", "error", err)
		return apperrors.ErrInvalidOtp
	}

	if time.Now().After(otpToken.ExpiresAt) || bookingId != otpToken.BookingId {
		slog.Error("invalid otp provided")
		return apperrors.ErrInvalidOtp
	}

	tx, err := s.bookingRepository.BeginTx(ctx)
	if err != nil {
		slog.Error("failed to start confirm pickup", "error", err)
		return err
	}

	defer func() {
		if txErr := s.bookingRepository.HandleTransaction(ctx, tx, err); txErr != nil {
			slog.Error("failed to handle transaction", "error", txErr)
			err = txErr
		}
	}()

	err = s.bookingRepository.UpdateBookingStatus(ctx, tx, bookingId, CheckedOut)
	if err != nil {
		slog.Error("failed to update booking status", "error", err)
		return err
	}

	err = s.bookingRepository.UpdateActualPickupTime(ctx, tx, bookingId)
	if err != nil {
		slog.Error("failed to update actual pickup time for booking", "error", err)
		return err
	}

	err = s.bookingRepository.DeleteOtpTokenById(ctx, nil, otpToken.Id)
	if err != nil {
		slog.Warn("failed to delete otp token", "error", err)
	}

	return nil
}

func (s *service) InitiateReturn(ctx context.Context, bookingId int) (err error) {
	userId, ok := ctx.Value(middleware.RequestContextUserIdKey).(int)
	if !ok {
		slog.Error("failed to retrieve user id from context")
		return apperrors.ErrInternalServer
	}

	booking, err := s.bookingRepository.GetBookingById(ctx, nil, bookingId)
	if err != nil {
		slog.Error("failed to fetch booking by id", "error", err)
		return err
	}

	if booking.HostId != userId || booking.Status != CheckedOut {
		slog.Error("invalid booking initiate return attempt")
		return apperrors.ErrActionForbidden
	}

	host, err := s.userService.GetUserById(ctx, userId)
	if err != nil {
		slog.Error("failed to get the user to initiate return", "error", err)
		return err
	}

	otp, err := cryptokit.GenerateOTP()
	if err != nil {
		slog.Error("failed to generate secure otp", "error", err)
		return apperrors.ErrInternalServer
	}

	tx, err := s.bookingRepository.BeginTx(ctx)
	if err != nil {
		slog.Error("failed to start initiate return", "error", err)
		return err
	}

	defer func() {
		if txErr := s.bookingRepository.HandleTransaction(ctx, tx, err); txErr != nil {
			slog.Error("failed to handle transaction", "error", txErr)
			err = txErr
		}
	}()

	optTokenData := OtpToken{
		BookingId: booking.Id,
		Otp:       otp,
		ExpiresAt: time.Now().Add(20 * time.Minute),
	}
	err = s.bookingRepository.CreateOtpToken(ctx, tx, repository.OtpToken(optTokenData))
	if err != nil {
		slog.Error("failed to create return otp token", "error", err)
		return err
	}

	emailBody := fmt.Sprintf(initiateReturnOtpEmailContent, host.Name, otp)

	err = s.emailService.SendEmail(host.Name, host.Email, "Vehicle Return OTP – Wheelio", emailBody)
	if err != nil {
		slog.Error("failed to send return otp email", "error", err)
		return err
	}

	return nil
}

func (s *service) ConfirmReturn(ctx context.Context, bookingId int, otpData OtpRequestBody) (err error) {
	userId, ok := ctx.Value(middleware.RequestContextUserIdKey).(int)
	if !ok {
		slog.Error("failed to retrieve user id from context")
		return apperrors.ErrInternalServer
	}

	booking, err := s.bookingRepository.GetBookingById(ctx, nil, bookingId)
	if err != nil {
		slog.Error("failed to fetch booking by id", "error", err)
		return err
	}

	if booking.SeekerId != userId || booking.Status != CheckedOut {
		slog.Error("invalid booking return attempt")
		return apperrors.ErrActionForbidden
	}

	otpToken, err := s.bookingRepository.GetOtpToken(ctx, nil, otpData.Otp)
	if err != nil {
		slog.Error("failed to get otp", "error", err)
		return apperrors.ErrInvalidOtp
	}

	if time.Now().After(otpToken.ExpiresAt) || bookingId != otpToken.BookingId {
		slog.Error("invalid otp provided")
		return apperrors.ErrInvalidOtp
	}

	tx, err := s.bookingRepository.BeginTx(ctx)
	if err != nil {
		slog.Error("failed to start confirm return", "error", err)
		return err
	}

	defer func() {
		if txErr := s.bookingRepository.HandleTransaction(ctx, tx, err); txErr != nil {
			slog.Error("failed to handle transaction", "error", txErr)
			err = txErr
		}
	}()

	err = s.bookingRepository.UpdateBookingStatus(ctx, tx, bookingId, Returned)
	if err != nil {
		slog.Error("failed to update booking status", "error", err)
		return err
	}

	err = s.bookingRepository.UpdateActualDropoffTime(ctx, tx, bookingId)
	if err != nil {
		slog.Error("failed to update actual dropoff time for booking", "error", err)
		return err
	}

	overdueTime := time.Since(booking.ScheduledDropoffTime).Hours()
	if overdueTime < 0 {
		overdueTime = 0
	}
	additionalFees := overdueTime * booking.OverdueFeeRatePerHour
	taxAmount := (booking.BookingAmount + additionalFees) * taxRate
	totalAmount := booking.BookingAmount + additionalFees + taxAmount

	invoiceData := repository.Invoice{
		BookingId:      bookingId,
		BookingAmount:  booking.BookingAmount,
		AdditionalFees: additionalFees,
		Tax:            taxAmount,
		TaxRate:        taxRate,
		TotalAmount:    totalAmount,
	}
	_, err = s.bookingRepository.CreateInvoice(ctx, tx, invoiceData)
	if err != nil {
		slog.Error("failed to create invoice", "error", err)
		return err
	}

	err = s.bookingRepository.DeleteOtpTokenById(ctx, nil, otpToken.Id)
	if err != nil {
		slog.Warn("failed to delete otp token", "error", err)
	}

	return nil
}

func (s *service) GetSeekerBookings(ctx context.Context, page, limit int) (bookings PaginatedBookingData, err error) {
	userId, ok := ctx.Value(middleware.RequestContextUserIdKey).(int)
	if !ok {
		slog.Error("failed to retrieve user id from context")
		return PaginatedBookingData{}, apperrors.ErrInternalServer
	}

	if page <= 0 {
		slog.Error("invalid page number provided", "page", page)
		return PaginatedBookingData{}, apperrors.ErrInvalidPagination
	}

	if limit <= 0 {
		slog.Error("invalid limit value provided", "limit", limit)
		return PaginatedBookingData{}, apperrors.ErrInvalidPagination
	}

	offset := limit * (page - 1)

	repoParams := repository.GetSeekerBookingsParams{
		SeekerId: userId,
		Offset:   offset,
		Limit:    limit,
	}
	bookingList, totalBookings, err := s.bookingRepository.GetSeekerBookings(ctx, nil, repoParams)
	if err != nil {
		slog.Error("failed to get vehicle list for host", "error", err)
		return PaginatedBookingData{}, err
	}

	vehicleData := make([]BookingData, len(bookingList))
	for i, v := range bookingList {
		vehicleData[i] = BookingData(v)
	}

	return PaginatedBookingData{
		Data: vehicleData,
		Pagination: PaginationParams{
			Page:       page,
			PageSize:   limit,
			TotalCount: totalBookings,
		}}, nil
}

func (s *service) GetHostBookings(ctx context.Context, page, limit int) (bookings PaginatedBookingData, err error) {
	userId, ok := ctx.Value(middleware.RequestContextUserIdKey).(int)
	if !ok {
		slog.Error("failed to retrieve user id from context")
		return PaginatedBookingData{}, apperrors.ErrInternalServer
	}

	if page <= 0 {
		slog.Error("invalid page number provided", "page", page)
		return PaginatedBookingData{}, apperrors.ErrInvalidPagination
	}

	if limit <= 0 {
		slog.Error("invalid limit value provided", "limit", limit)
		return PaginatedBookingData{}, apperrors.ErrInvalidPagination
	}

	offset := limit * (page - 1)

	repoParams := repository.GetHostBookingsParams{
		HostId: userId,
		Offset: offset,
		Limit:  limit,
	}
	bookingList, totalBookings, err := s.bookingRepository.GetHostBookings(ctx, nil, repoParams)
	if err != nil {
		slog.Error("failed to get vehicle list for host", "error", err)
		return PaginatedBookingData{}, err
	}

	vehicleData := make([]BookingData, len(bookingList))
	for i, v := range bookingList {
		vehicleData[i] = BookingData(v)
	}

	return PaginatedBookingData{
		Data: vehicleData,
		Pagination: PaginationParams{
			Page:       page,
			PageSize:   limit,
			TotalCount: totalBookings,
		}}, nil
}

func (s *service) GetBookingDetailsById(ctx context.Context, bookingId int) (booking BookingDetails, err error) {
	bookingDetails, err := s.bookingRepository.GetBookingDetailsById(ctx, nil, bookingId)
	if err != nil {
		slog.Error("failed to get booking details", "error", err)
		return BookingDetails{}, err
	}

	return mapBookingDetailsRepoToBookingDetails(bookingDetails), nil
}
