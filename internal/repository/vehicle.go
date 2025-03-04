package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
)

type vehicleRepository struct {
	BaseRepository
}

type VehicleRepository interface {
	RepositoryTransaction
	CreateVehicle(ctx context.Context, tx *sql.Tx, vehicleData CreateVehicleRequestBody) (Vehicle, error)
	UpdateVehicle(ctx context.Context, tx *sql.Tx, vehicleData EditVehicleRequestBody) (Vehicle, error)
	SoftDeleteVehicle(ctx context.Context, tx *sql.Tx, vehicleId int) error
	CreateVehicleImage(ctx context.Context, tx *sql.Tx, vehicleImageData CreateVehicleImageData) (VehicleImage, error)
	DeleteAllImagesForVehicle(ctx context.Context, tx *sql.Tx, vehicleId int) error
	GetVehicleById(ctx context.Context, tx *sql.Tx, vehicleId int) (Vehicle, error)
	GetVehicleImagesByVehicleId(ctx context.Context, tx *sql.Tx, vehicleId int) ([]VehicleImage, error)
	GetVehicles(ctx context.Context, tx *sql.Tx, params GetVehiclesParams) ([]VehicleOverview, int, error)
	GetVehiclesForHost(ctx context.Context, tx *sql.Tx, params GetVehiclesForHostParams) ([]VehicleOverview, int, error)
}

func NewVehicleRepository(db *sql.DB) VehicleRepository {
	return &vehicleRepository{
		BaseRepository: BaseRepository{db},
	}
}

const (
	createVehicleQuery = `
	INSERT INTO vehicles (
		name, 
		fuel_type, 
		seat_count, 
		transmission_type, 
		features, 
		rate_per_hour, 
		overdue_fee_rate_per_hour, 
		address, 
		state, 
		city, 
		pin_code, 
		cancellation_allowed, 
		host_id
	) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) 
	RETURNING *;`

	updateVehicleQuery = `
	UPDATE vehicles 
	SET 
		name = $1, 
		fuel_type = $2, 
		seat_count = $3, 
		transmission_type = $4, 
		features = $5, 
		rate_per_hour = $6, 
		overdue_fee_rate_per_hour = $7, 
		address = $8, 
		state = $9, 
		city = $10, 
		pin_code = $11, 
		cancellation_allowed = $12 
	WHERE id = $13 AND is_deleted=false
	RETURNING *;`

	softDeleteVehicleQuery = "UPDATE vehicles SET is_deleted=true WHERE id=$1"

	createVehicleImageQuery = `
	INSERT INTO vehicle_images (
		vehicle_id,
		url,
		featured
	)
	VALUES ($1, $2, $3) 
	RETURNING *;`

	deleteAllImagesForVehicleQuery = "DELETE FROM vehicle_images WHERE vehicle_id=$1"

	getVehicleByIdQuery = "SELECT * FROM vehicles WHERE id=$1 AND is_deleted=false"

	getVehicleImagesByVehicleIdQuery = "SELECT * FROM vehicle_images WHERE vehicle_id=$1"

	getVehiclesQuery = `
	SELECT 
		v.id,
		v.name,
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
		v.rate_per_hour,
		v.address,
		v.pin_code,
		COUNT(*) OVER() AS total_count
	FROM vehicles v
	WHERE 
		v.city ILIKE $1 AND
		v.is_deleted = false AND 
		v.available = true AND
			NOT EXISTS (
			SELECT 1
			FROM bookings AS b
			WHERE
				v.id = b.vehicle_id AND
				b.status NOT IN ('RETURNED', 'CANCELLED') AND
				(
					(b.scheduled_pickup_time <= $2 AND $2 <= b.scheduled_dropoff_time) OR
					(b.scheduled_pickup_time <= $3 AND $3 <= b.scheduled_dropoff_time) OR
					($2 <= b.scheduled_pickup_time AND b.scheduled_dropoff_time <= $3) 
				)
			)
	ORDER BY v.created_at DESC
	OFFSET $4
	LIMIT $5;`

	getVehiclesForHostQuery = `
	SELECT 
		v.id,
		v.name,
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
		v.rate_per_hour,
		v.address,
		v.pin_code,
		COUNT(*) OVER() AS total_count
	FROM vehicles v
	WHERE 
		host_id=$1 AND
		v.is_deleted = false
	ORDER BY v.created_at DESC
	OFFSET $2
	LIMIT $3;`
)

func (vr *vehicleRepository) CreateVehicle(ctx context.Context, tx *sql.Tx, vehicleData CreateVehicleRequestBody) (Vehicle, error) {
	executer := vr.initiateQueryExecuter(tx)

	var vehicle Vehicle
	err := executer.QueryRowContext(
		ctx,
		createVehicleQuery,
		vehicleData.Name,
		vehicleData.FuelType,
		vehicleData.SeatCount,
		vehicleData.TransmissionType,
		vehicleData.Features,
		vehicleData.RatePerHour,
		vehicleData.OverdueFeeRatePerHour,
		vehicleData.Address,
		vehicleData.State,
		vehicleData.City,
		vehicleData.PinCode,
		vehicleData.CancellationAllowed,
		vehicleData.HostId,
	).Scan(
		&vehicle.Id,
		&vehicle.Name,
		&vehicle.FuelType,
		&vehicle.SeatCount,
		&vehicle.TransmissionType,
		&vehicle.Features,
		&vehicle.RatePerHour,
		&vehicle.OverdueFeeRatePerHour,
		&vehicle.Address,
		&vehicle.State,
		&vehicle.City,
		&vehicle.PinCode,
		&vehicle.CancellationAllowed,
		&vehicle.Available,
		&vehicle.HostId,
		&vehicle.IsDeleted,
		&vehicle.CreatedAt,
		&vehicle.UpdatedAt,
	)
	if err != nil {
		slog.Error("failed to create vehicle", "error", err)
		return Vehicle{}, apperrors.ErrInternalServer
	}

	return vehicle, nil
}

func (vr *vehicleRepository) UpdateVehicle(ctx context.Context, tx *sql.Tx, vehicleData EditVehicleRequestBody) (Vehicle, error) {
	executer := vr.initiateQueryExecuter(tx)

	var vehicle Vehicle
	err := executer.QueryRowContext(
		ctx,
		updateVehicleQuery,
		vehicleData.Name,
		vehicleData.FuelType,
		vehicleData.SeatCount,
		vehicleData.TransmissionType,
		vehicleData.Features,
		vehicleData.RatePerHour,
		vehicleData.OverdueFeeRatePerHour,
		vehicleData.Address,
		vehicleData.State,
		vehicleData.City,
		vehicleData.PinCode,
		vehicleData.CancellationAllowed,
		vehicleData.Id,
	).Scan(
		&vehicle.Id,
		&vehicle.Name,
		&vehicle.FuelType,
		&vehicle.SeatCount,
		&vehicle.TransmissionType,
		&vehicle.Features,
		&vehicle.RatePerHour,
		&vehicle.OverdueFeeRatePerHour,
		&vehicle.Address,
		&vehicle.State,
		&vehicle.City,
		&vehicle.PinCode,
		&vehicle.CancellationAllowed,
		&vehicle.Available,
		&vehicle.HostId,
		&vehicle.IsDeleted,
		&vehicle.CreatedAt,
		&vehicle.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("no vehicle found", "error", err)
			return Vehicle{}, apperrors.ErrVehicleNotFound
		}
		slog.Error("failed to update vehicle", "error", err)
		return Vehicle{}, err
	}

	return vehicle, nil
}

func (vr *vehicleRepository) SoftDeleteVehicle(ctx context.Context, tx *sql.Tx, vehicleId int) error {
	executer := vr.initiateQueryExecuter(tx)

	_, err := executer.ExecContext(ctx, softDeleteVehicleQuery, vehicleId)
	if err != nil {
		slog.Error("failed to soft delete vehicle", "error", err)
		return apperrors.ErrInternalServer
	}

	return nil
}

func (vr *vehicleRepository) CreateVehicleImage(ctx context.Context, tx *sql.Tx, vehicleImageData CreateVehicleImageData) (VehicleImage, error) {
	executer := vr.initiateQueryExecuter(tx)

	var vehicleImage VehicleImage
	err := executer.QueryRowContext(
		ctx,
		createVehicleImageQuery,
		vehicleImageData.VehicleId,
		vehicleImageData.Url,
		vehicleImageData.Featured,
	).Scan(
		&vehicleImage.Id,
		&vehicleImage.VehicleId,
		&vehicleImage.Url,
		&vehicleImage.Featured,
		&vehicleImage.CreatedAt,
	)
	if err != nil {
		slog.Error("failed to create vehicle image", "error", err)
		return VehicleImage{}, apperrors.ErrInternalServer
	}

	return vehicleImage, nil
}

func (vr *vehicleRepository) DeleteAllImagesForVehicle(ctx context.Context, tx *sql.Tx, vehicleId int) error {
	executer := vr.initiateQueryExecuter(tx)

	_, err := executer.ExecContext(ctx, deleteAllImagesForVehicleQuery, vehicleId)
	if err != nil {
		slog.Error("failed to delete images for vehicle", "error", err)
		return apperrors.ErrInternalServer
	}

	return nil
}

func (vr *vehicleRepository) GetVehicleById(ctx context.Context, tx *sql.Tx, vehicleId int) (Vehicle, error) {
	executer := vr.initiateQueryExecuter(tx)

	var vehicle Vehicle
	err := executer.QueryRowContext(
		ctx,
		getVehicleByIdQuery,
		vehicleId,
	).Scan(
		&vehicle.Id,
		&vehicle.Name,
		&vehicle.FuelType,
		&vehicle.SeatCount,
		&vehicle.TransmissionType,
		&vehicle.Features,
		&vehicle.RatePerHour,
		&vehicle.OverdueFeeRatePerHour,
		&vehicle.Address,
		&vehicle.State,
		&vehicle.City,
		&vehicle.PinCode,
		&vehicle.CancellationAllowed,
		&vehicle.Available,
		&vehicle.HostId,
		&vehicle.IsDeleted,
		&vehicle.CreatedAt,
		&vehicle.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("no vehicle found", "error", err)
			return Vehicle{}, apperrors.ErrVehicleNotFound
		}
		slog.Error("failed to get vehicle", "error", err)
		return Vehicle{}, apperrors.ErrInternalServer
	}

	return vehicle, nil
}

func (vr *vehicleRepository) GetVehicleImagesByVehicleId(ctx context.Context, tx *sql.Tx, vehicleId int) ([]VehicleImage, error) {
	executer := vr.initiateQueryExecuter(tx)

	var vehicleImages []VehicleImage
	rows, err := executer.QueryContext(ctx, getVehicleImagesByVehicleIdQuery, vehicleId)
	if err != nil {
		slog.Error("failed to get vehicle images", "error", err)
		return []VehicleImage{}, apperrors.ErrInternalServer
	}

	defer rows.Close()
	for rows.Next() {
		var vehicleImage VehicleImage
		err = rows.Scan(&vehicleImage.Id, &vehicleImage.VehicleId, &vehicleImage.Url, &vehicleImage.Featured, &vehicleImage.CreatedAt)
		if err != nil {
			slog.Error("failed to scan vehicle image from rows", "error", err)
			return []VehicleImage{}, apperrors.ErrInternalServer
		}
		vehicleImages = append(vehicleImages, vehicleImage)
	}

	err = rows.Err()
	if err != nil {
		slog.Error("failed iterate over vehicle image rows", "error", err)
		return []VehicleImage{}, apperrors.ErrInternalServer
	}
	return vehicleImages, nil
}

func (vr *vehicleRepository) GetVehicles(ctx context.Context, tx *sql.Tx, params GetVehiclesParams) ([]VehicleOverview, int, error) {
	executer := vr.initiateQueryExecuter(tx)

	var vehicles []VehicleOverview
	var totalCount int
	rows, err := executer.QueryContext(
		ctx,
		getVehiclesQuery,
		params.City,
		params.PickupTimestamp,
		params.DropoffTimestamp,
		params.Offset,
		params.Limit,
	)
	if err != nil {
		slog.Error("failed to get vehicles", "error", err)
		return []VehicleOverview{}, 0, apperrors.ErrInternalServer
	}

	defer rows.Close()
	for rows.Next() {
		var vehicleData VehicleOverview
		err = rows.Scan(
			&vehicleData.Id,
			&vehicleData.Name,
			&vehicleData.FuelType,
			&vehicleData.SeatCount,
			&vehicleData.TransmissionType,
			&vehicleData.Image,
			&vehicleData.RatePerHour,
			&vehicleData.Address,
			&vehicleData.PinCode,
			&totalCount,
		)
		if err != nil {
			slog.Error("failed to scan vehicle overview from rows", "error", err)
			return []VehicleOverview{}, 0, apperrors.ErrInternalServer
		}
		vehicles = append(vehicles, vehicleData)
	}

	err = rows.Err()
	if err != nil {
		slog.Error("failed iterate over vehicle overview rows", "error", err)
		return []VehicleOverview{}, 0, apperrors.ErrInternalServer
	}
	return vehicles, totalCount, nil
}

func (vr *vehicleRepository) GetVehiclesForHost(ctx context.Context, tx *sql.Tx, params GetVehiclesForHostParams) ([]VehicleOverview, int, error) {
	executer := vr.initiateQueryExecuter(tx)

	var vehicles []VehicleOverview
	var totalCount int
	rows, err := executer.QueryContext(
		ctx,
		getVehiclesForHostQuery,
		params.HostId,
		params.Offset,
		params.Limit,
	)
	if err != nil {
		slog.Error("failed to get vehicles for host", "error", err)
		return []VehicleOverview{}, 0, apperrors.ErrInternalServer
	}

	defer rows.Close()
	for rows.Next() {
		var vehicleData VehicleOverview
		err = rows.Scan(
			&vehicleData.Id,
			&vehicleData.Name,
			&vehicleData.FuelType,
			&vehicleData.SeatCount,
			&vehicleData.TransmissionType,
			&vehicleData.Image,
			&vehicleData.RatePerHour,
			&vehicleData.Address,
			&vehicleData.PinCode,
			&totalCount,
		)
		if err != nil {
			slog.Error("failed to scan vehicle overview from rows for host", "error", err)
			return []VehicleOverview{}, 0, apperrors.ErrInternalServer
		}
		vehicles = append(vehicles, vehicleData)
	}

	err = rows.Err()
	if err != nil {
		slog.Error("failed iterate over vehicle overview rows for host", "error", err)
		return []VehicleOverview{}, 0, apperrors.ErrInternalServer
	}
	return vehicles, totalCount, nil
}
