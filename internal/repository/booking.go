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
	GetOtpToken(ctx context.Context, tx *sql.Tx, otp string) (OtpToken, error)
	DeleteOtpTokenById(ctx context.Context, tx *sql.Tx, otpTokenId int) error
	UpdateBookingStatus(ctx context.Context, tx *sql.Tx, bookingId int, status string) error
	UpdateActualPickupTime(ctx context.Context, tx *sql.Tx, bookingId int) error
	UpdateActualDropoffTime(ctx context.Context, tx *sql.Tx, bookingId int) error
	GetBookingById(ctx context.Context, tx *sql.Tx, bookingId int) (Booking, error)
	CreateInvoice(ctx context.Context, tx *sql.Tx, invoiceData Invoice) (Invoice, error)
	GetSeekerBookings(ctx context.Context, tx *sql.Tx, params GetSeekerBookingsParams) ([]BookingData, int, error)
	GetHostBookings(ctx context.Context, tx *sql.Tx, params GetHostBookingsParams) ([]BookingData, int, error)
	GetBookingDetailsById(ctx context.Context, tx *sql.Tx, bookingId int) (BookingDetails, error)
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
		cancellation_allowed,
		scheduled_pickup_time,
		scheduled_dropoff_time
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
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

	getOtpTokenQuery = "SELECT * FROM otp_tokens WHERE otp=$1;"

	deleteOtpTokenByIdQuery = "DELETE FROM otp_tokens WHERE id=$1;"

	updateBookingStatusQuery = "UPDATE bookings SET status=$1 WHERE id=$2;"

	getBookingById = "SELECT * FROM bookings WHERE id=$1"

	createInvoiceQuery = `
	INSERT INTO invoices (
		booking_id,
		booking_amount,
		additional_fees,
		tax,
		tax_rate,
		total_amount
	) VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING *;`

	updateActualPickupTimeQuery = `
	UPDATE bookings
	SET actual_pickup_time = CURRENT_TIMESTAMP
	WHERE id = $1;`

	updateActualDropoffTimeQuery = `
	UPDATE bookings
	SET actual_dropoff_time = CURRENT_TIMESTAMP
	WHERE id = $1;`

	getSeekerBookingsQuery = `
	SELECT 
		b.id,
		b.status,
		b.pickup_location,
		b.dropoff_location,
		b.booking_amount,
		b.overdue_fee_rate_per_hour,
		b.cancellation_allowed,
		b.scheduled_pickup_time,
		b.scheduled_dropoff_time,
		v.name AS vehicleName,
		v.seat_count AS vehicleSeatCount,
		v.fuel_type AS vehicleFuelType,
		v.transmission_type AS vehicleTransmissionType,
		COALESCE((
			SELECT vi.url
			FROM vehicle_images vi
			WHERE vi.vehicle_id = v.id
			AND vi.featured = true
			LIMIT 1
		), '') AS vehicleImage,
		COUNT(*) OVER() AS total_count
	FROM bookings b
	JOIN vehicles v ON b.vehicle_id = v.id
	WHERE b.seeker_id=$1
	ORDER BY b.created_at DESC
	OFFSET $2
	LIMIT $3;`

	getHostBookingsQuery = `
	SELECT 
		b.id,
		b.status,
		b.pickup_location,
		b.dropoff_location,
		b.booking_amount,
		b.overdue_fee_rate_per_hour,
		b.cancellation_allowed,
		b.scheduled_pickup_time,
		b.scheduled_dropoff_time,
		v.name AS vehicleName,
		v.seat_count AS vehicleSeatCount,
		v.fuel_type AS vehicleFuelType,
		v.transmission_type AS vehicleTransmissionType,
		COALESCE((
			SELECT vi.url
			FROM vehicle_images vi
			WHERE vi.vehicle_id = v.id
			AND vi.featured = true
			LIMIT 1
		), '') AS vehicleImage,
		COUNT(*) OVER() AS total_count
	FROM bookings b
	JOIN vehicles v ON b.vehicle_id = v.id
	WHERE b.host_id=$1
	ORDER BY b.created_at DESC
	OFFSET $2
	LIMIT $3;`

	getBookingDetailsByIdQuery = `
	SELECT 
		b.id AS booking_id,
		b.status,
		b.pickup_location,
		b.dropoff_location,
		b.booking_amount,
		b.overdue_fee_rate_per_hour,
		b.cancellation_allowed,
		b.actual_pickup_time,
		b.actual_dropoff_time,
		b.scheduled_pickup_time,
		b.scheduled_dropoff_time,
		h.id AS host_id,
		h.name AS host_name,
		h.email AS host_email,
		h.phone_number AS host_phone,
		s.id AS seeker_id,
		s.name AS seeker_name,
		s.email AS seeker_email,
		s.phone_number AS seeker_phone,
		v.id AS vehicle_id,
		v.name AS vehicle_name,
		v.fuel_type,
		v.seat_count,
		v.transmission_type,
		COALESCE((
			SELECT vi.url
			FROM vehicle_images vi
			WHERE vi.vehicle_id = v.id
			AND vi.featured = true
			LIMIT 1
		), '') AS image,
		COALESCE(i.id, 0) AS invoice_id,
		COALESCE(i.additional_fees, 0) AS additional_fees,
		COALESCE(i.tax, 0) AS tax,
		COALESCE(i.tax_rate, 0) AS tax_rate,
		COALESCE(i.total_amount, 0) AS total_amount
	FROM bookings b
	JOIN users h ON b.host_id = h.id
	JOIN users s ON b.seeker_id = s.id
	JOIN vehicles v ON b.vehicle_id = v.id
	LEFT JOIN invoices i ON i.booking_id = b.id
	WHERE b.id = $1;`
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
		bookingData.CancellationAllowed,
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
		&booking.CancellationAllowed,
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

func (br *bookingRepository) GetOtpToken(ctx context.Context, tx *sql.Tx, otp string) (OtpToken, error) {
	executer := br.initiateQueryExecuter(tx)

	var optToken OtpToken
	err := executer.QueryRowContext(
		ctx,
		getOtpTokenQuery,
		otp,
	).Scan(
		&optToken.Id,
		&optToken.BookingId,
		&optToken.Otp,
		&optToken.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return OtpToken{}, apperrors.ErrOptTokenNotFound
		}
		return OtpToken{}, apperrors.ErrInternalServer
	}

	return optToken, nil
}

func (br *bookingRepository) DeleteOtpTokenById(ctx context.Context, tx *sql.Tx, otpTokenId int) error {
	executer := br.initiateQueryExecuter(tx)

	_, err := executer.ExecContext(ctx, deleteOtpTokenByIdQuery, otpTokenId)
	if err != nil {
		slog.Error("failed to delete opt token", "error", err)
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

func (br *bookingRepository) UpdateActualPickupTime(ctx context.Context, tx *sql.Tx, bookingId int) error {
	executer := br.initiateQueryExecuter(tx)

	_, err := executer.ExecContext(ctx, updateActualPickupTimeQuery, bookingId)
	if err != nil {
		slog.Error("failed to update actual pickup time for vehicle", "error", err)
		return apperrors.ErrInternalServer
	}

	return nil
}

func (br *bookingRepository) UpdateActualDropoffTime(ctx context.Context, tx *sql.Tx, bookingId int) error {
	executer := br.initiateQueryExecuter(tx)

	_, err := executer.ExecContext(ctx, updateActualDropoffTimeQuery, bookingId)
	if err != nil {
		slog.Error("failed to update actual dropoff time for vehicle", "error", err)
		return apperrors.ErrInternalServer
	}

	return nil
}

func (br *bookingRepository) GetBookingById(ctx context.Context, tx *sql.Tx, bookingId int) (Booking, error) {
	executer := br.initiateQueryExecuter(tx)

	var booking Booking
	err := executer.QueryRowContext(
		ctx,
		getBookingById,
		bookingId,
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
		&booking.CancellationAllowed,
		&booking.ActualPickupTime,
		&booking.ActualDropoffTime,
		&booking.ScheduledPickupTime,
		&booking.ScheduledDropoffTime,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Booking{}, apperrors.ErrBookingNotFound
		}
		return Booking{}, apperrors.ErrInternalServer
	}

	return booking, nil
}

func (br *bookingRepository) CreateInvoice(ctx context.Context, tx *sql.Tx, invoiceData Invoice) (Invoice, error) {
	executer := br.initiateQueryExecuter(tx)

	var invoice Invoice
	err := executer.QueryRowContext(
		ctx,
		createInvoiceQuery,
		invoiceData.BookingId,
		invoiceData.BookingAmount,
		invoiceData.AdditionalFees,
		invoiceData.Tax,
		invoiceData.TaxRate,
		invoiceData.TotalAmount,
	).Scan(
		&invoice.Id,
		&invoice.BookingId,
		&invoice.BookingAmount,
		&invoice.AdditionalFees,
		&invoice.Tax,
		&invoice.TaxRate,
		&invoice.TotalAmount,
	)
	if err != nil {
		slog.Error("failed to create invoice", "error", err)
		return Invoice{}, apperrors.ErrInternalServer
	}

	return invoice, nil
}

func (br *bookingRepository) GetSeekerBookings(ctx context.Context, tx *sql.Tx, params GetSeekerBookingsParams) ([]BookingData, int, error) {
	executer := br.initiateQueryExecuter(tx)

	var bookings []BookingData
	var totalCount int
	rows, err := executer.QueryContext(
		ctx,
		getSeekerBookingsQuery,
		params.SeekerId,
		params.Offset,
		params.Limit,
	)
	if err != nil {
		slog.Error("failed to get bookings for seeker", "error", err)
		return []BookingData{}, 0, apperrors.ErrInternalServer
	}

	defer rows.Close()
	for rows.Next() {
		var bookingData BookingData
		err = rows.Scan(
			&bookingData.Id,
			&bookingData.Status,
			&bookingData.PickupLocation,
			&bookingData.DropoffLocation,
			&bookingData.BookingAmount,
			&bookingData.OverdueFeeRatePerHour,
			&bookingData.CancellationAllowed,
			&bookingData.ScheduledPickupTime,
			&bookingData.ScheduledDropoffTime,
			&bookingData.VehicleName,
			&bookingData.VehicleSeatCount,
			&bookingData.VehicleFuelType,
			&bookingData.VehicleTransmissionType,
			&bookingData.VehicleImage,
			&totalCount,
		)
		if err != nil {
			slog.Error("failed to scan booking data from rows for seeker", "error", err)
			return []BookingData{}, 0, apperrors.ErrInternalServer
		}
		bookings = append(bookings, bookingData)
	}

	err = rows.Err()
	if err != nil {
		slog.Error("failed iterate over booking data rows for seeker", "error", err)
		return []BookingData{}, 0, apperrors.ErrInternalServer
	}

	return bookings, totalCount, nil
}

func (br *bookingRepository) GetHostBookings(ctx context.Context, tx *sql.Tx, params GetHostBookingsParams) ([]BookingData, int, error) {
	executer := br.initiateQueryExecuter(tx)

	var bookings []BookingData
	var totalCount int
	rows, err := executer.QueryContext(
		ctx,
		getHostBookingsQuery,
		params.HostId,
		params.Offset,
		params.Limit,
	)
	if err != nil {
		slog.Error("failed to get bookings for host", "error", err)
		return []BookingData{}, 0, apperrors.ErrInternalServer
	}

	defer rows.Close()
	for rows.Next() {
		var bookingData BookingData
		err = rows.Scan(
			&bookingData.Id,
			&bookingData.Status,
			&bookingData.PickupLocation,
			&bookingData.DropoffLocation,
			&bookingData.BookingAmount,
			&bookingData.OverdueFeeRatePerHour,
			&bookingData.CancellationAllowed,
			&bookingData.ScheduledPickupTime,
			&bookingData.ScheduledDropoffTime,
			&bookingData.VehicleName,
			&bookingData.VehicleSeatCount,
			&bookingData.VehicleFuelType,
			&bookingData.VehicleTransmissionType,
			&bookingData.VehicleImage,
			&totalCount,
		)
		if err != nil {
			slog.Error("failed to scan booking data from rows for host", "error", err)
			return []BookingData{}, 0, apperrors.ErrInternalServer
		}
		bookings = append(bookings, bookingData)
	}

	err = rows.Err()
	if err != nil {
		slog.Error("failed iterate over booking data rows for host", "error", err)
		return []BookingData{}, 0, apperrors.ErrInternalServer
	}

	return bookings, totalCount, nil
}

func (br *bookingRepository) GetBookingDetailsById(ctx context.Context, tx *sql.Tx, bookingId int) (BookingDetails, error) {
	executer := br.initiateQueryExecuter(tx)

	var booking BookingDetails
	var host BookingDetailsUser
	var seeker BookingDetailsUser
	var vehicle BookingDetailsVehicle
	var invoice BookingDetailsInvoice
	err := executer.QueryRowContext(
		ctx,
		getBookingDetailsByIdQuery,
		bookingId,
	).Scan(
		&booking.Id,
		&booking.Status,
		&booking.PickupLocation,
		&booking.DropoffLocation,
		&booking.BookingAmount,
		&booking.OverdueFeeRatePerHour,
		&booking.CancellationAllowed,
		&booking.ActualPickupTime,
		&booking.ActualDropoffTime,
		&booking.ScheduledPickupTime,
		&booking.ScheduledDropoffTime,
		&host.Id,
		&host.Name,
		&host.Email,
		&host.PhoneNumber,
		&seeker.Id,
		&seeker.Name,
		&seeker.Email,
		&seeker.PhoneNumber,
		&vehicle.Id,
		&vehicle.Name,
		&vehicle.FuelType,
		&vehicle.SeatCount,
		&vehicle.TransmissionType,
		&vehicle.Image,
		&invoice.Id,
		&invoice.AdditionalFees,
		&invoice.Tax,
		&invoice.TaxRate,
		&invoice.TotalAmount,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("no booking found", "error", err)
			return BookingDetails{}, apperrors.ErrBookingNotFound
		}
		slog.Error("failed to fetch booking details", "error", err)
		return BookingDetails{}, apperrors.ErrInternalServer
	}

	booking.Host = host
	booking.Seeker = seeker
	booking.Vehicle = vehicle
	booking.Invoice = invoice

	return booking, nil
}
