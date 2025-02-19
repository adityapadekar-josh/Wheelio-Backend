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
