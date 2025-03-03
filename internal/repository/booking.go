package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
)

type bookingRepository struct {
	BaseRepository
}

type BookingRepository interface {
	RepositoryTransaction
	CreateBooking(ctx context.Context, tx *sql.Tx, bookingData CreateBookingRequestBody) (Booking, error)
	VehicleBookingConflictCheck(ctx context.Context, tx *sql.Tx, vehicleId int, scheduledPickupTimestamp, scheduledDropoffTimestamp time.Time) error
	CreateOtpToken(ctx context.Context, tx *sql.Tx, tokenData OtpToken) error
	UpdateBookingStatus(ctx context.Context, tx *sql.Tx, bookingId int, status string) error
}

func NewBookingRepository(db *sql.DB) BookingRepository {
	return &bookingRepository{
		BaseRepository: BaseRepository{db},
	}
}

const (
	createBookingQuery = `
	INSERT INTO bookings (
		vehicle_id,
		host_id,
		seeker_id,
		status,
		pickup_location,
		dropoff_location,
		booking_amount,
		overdue_fee_rate_per_hour,
		scheduled_pickup_time,
		scheduled_dropoff_time
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING *;`

	vehicleBookingConflictCheckQuery = `
	SELECT 1
	FROM bookings
	WHERE
		vehicle_id=$1 AND
		status NOT IN ('RETURNED', 'CANCELLED') AND
		(
			(scheduled_pickup_time <= $2 AND $2 <= scheduled_dropoff_time) OR
			(scheduled_pickup_time <= $3 AND $3 <= scheduled_dropoff_time) OR
			($2 <= scheduled_pickup_time AND scheduled_dropoff_time <= $3) 
		)
	LIMIT 1;`

	createOtpTokenQuery = `
	INSERT INTO otp_tokens (
		booking_id,
		otp,
		expires_at
	) VALUES ($1, $2, $3)
	RETURNING *;`

	updateBookingStatusQuery = `UPDATE bookings SET status=$1 WHERE id=$2`
)

func (br *bookingRepository) CreateBooking(ctx context.Context, tx *sql.Tx, bookingData CreateBookingRequestBody) (Booking, error) {
	executer := br.initiateQueryExecuter(tx)

	var booking Booking
	err := executer.QueryRowContext(
		ctx,
		createBookingQuery,
		bookingData.VehicleId,
		bookingData.HostId,
		bookingData.SeekerId,
		bookingData.Status,
		bookingData.PickupLocation,
		bookingData.DropoffLocation,
		bookingData.BookingAmount,
		bookingData.OverdueFeeRatePerHour,
		bookingData.ScheduledPickupTime,
		bookingData.ScheduledDropoffTime,
	).Scan(
		&booking.Id,
		&booking.VehicleId,
		&booking.HostId,
		&booking.SeekerId,
		&booking.Status,
		&booking.PickupLocation,
		&booking.DropoffLocation,
		&booking.BookingAmount,
		&booking.OverdueFeeRatePerHour,
		&booking.ActualPickupTime,
		&booking.ActualDropoffTime,
		&booking.ScheduledPickupTime,
		&booking.ScheduledDropoffTime,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)
	if err != nil {
		slog.Error("failed to create booking", "error", err)
		return Booking{}, apperrors.ErrInternalServer
	}

	return booking, nil
}

func (br *bookingRepository) VehicleBookingConflictCheck(ctx context.Context, tx *sql.Tx, vehicleId int, scheduledPickupTimestamp, scheduledDropoffTimestamp time.Time) error {
	executer := br.initiateQueryExecuter(tx)

	var flag int
	err := executer.QueryRowContext(
		ctx,
		vehicleBookingConflictCheckQuery,
		vehicleId,
		scheduledPickupTimestamp,
		scheduledDropoffTimestamp,
	).Scan(&flag)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		slog.Error("failed to check vehicle booking conflict", "error", err)
		return apperrors.ErrInternalServer
	}

	return apperrors.ErrBookingConflict
}

func (br *bookingRepository) CreateOtpToken(ctx context.Context, tx *sql.Tx, tokenData OtpToken) error {
	executer := br.initiateQueryExecuter(tx)

	_, err := executer.ExecContext(
		ctx,
		createOtpTokenQuery,
		tokenData.BookingId,
		tokenData.Otp,
		tokenData.ExpiresAt,
	)
	if err != nil {
		slog.Error("failed to create otp token", "error", err)
		return apperrors.ErrInternalServer
	}

	return nil
}

func (br *bookingRepository) UpdateBookingStatus(ctx context.Context, tx *sql.Tx, bookingId int, status string) error {
	executer := br.initiateQueryExecuter(tx)

	_, err := executer.ExecContext(ctx, updateBookingStatusQuery, status, bookingId)
	if err != nil {
		slog.Error("failed to update booking status", "error", err)
		return apperrors.ErrInternalServer
	}

	return nil
}
